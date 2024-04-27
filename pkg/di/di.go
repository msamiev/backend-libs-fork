package di

import (
	"errors"
	"fmt"
)

type (
	Container struct {
		services map[string]service
		deinit   []service
		errors   []error
	}
	service interface {
		Init() error
		Deinit() error
	}
	serviceImpl[T any] struct {
		init   func() (T, error)
		deinit func(T) error

		noReuse bool
		val     T
	}
)

func (s *serviceImpl[T]) Init() (err error) {
	if s.init == nil {
		return
	}

	if s.val, err = s.init(); err != nil {
		return
	}

	if !s.noReuse {
		s.init = nil
	}

	return
}

func (s *serviceImpl[T]) Deinit() (err error) {
	if s.deinit != nil {
		err = s.deinit(s.val)
		s.deinit = nil // fix multiple deinit
	}

	return
}

func New() *Container {
	return &Container{
		services: make(map[string]service),
		errors:   make([]error, 0),
		deinit:   make([]service, 0),
	}
}

func (c *Container) addErr(err error) {
	c.errors = append(c.errors, err)
}

func (c *Container) addDeinit(svc service) {
	c.deinit = append(c.deinit, svc)
}

// Release will call deinits in opposite order
// as it was initialized.
func (c *Container) Release() error {
	for i := len(c.deinit) - 1; i >= 0; i-- {
		if err := c.deinit[i].Deinit(); err != nil {
			c.addErr(err)
		}
	}

	if len(c.errors) == 0 {
		return nil
	}

	var errs string
	for i, err := range c.errors {
		errs += err.Error()
		if i+1 < len(c.errors) {
			errs += ": "
		}
	}

	return errors.New(errs)
}

func empty[T any]() (t T) { return }

func generateSvcName[T any](name string) string {
	svcName := fmt.Sprintf("%T", empty[T]())
	if svcName == "<nil>" {
		svcName = fmt.Sprintf("%T", new(T))
	}

	return fmt.Sprintf("%s<%s>", name, svcName)
}

func Set[T any](c *Container, opts ...func(*serviceImpl[T])) {
	SetNamed(c, "", opts...)
}

func SetNamed[T any](c *Container, name string, opts ...func(*serviceImpl[T])) {
	svcName := generateSvcName[*serviceImpl[T]](name)
	svc, ok := c.services[svcName].(*serviceImpl[T])
	if !ok {
		svc = new(serviceImpl[T])
	}

	for _, opt := range opts {
		opt(svc)
	}

	c.services[svcName] = svc
}

func Get[T any](c *Container) T {
	return GetNamed[T](c, "")
}

func GetNamed[T any](c *Container, name string) T {
	svcName := generateSvcName[*serviceImpl[T]](name)
	svc, ok := c.services[svcName]
	if !ok {
		err := fmt.Errorf("dependency not found: %s", svcName)
		c.addErr(err)
		panic(err.Error())
	}

	if err := svc.Init(); err != nil {
		err := fmt.Errorf("init dependency %s: %w", svcName, err)
		c.addErr(err)
		panic(err.Error())
	}

	c.addDeinit(svc)

	return svc.(*serviceImpl[T]).val //nolint:forcetypeassert // we may panic here
}

func OptInit[T any](f func() (T, error)) func(*serviceImpl[T]) {
	return func(s *serviceImpl[T]) { s.init = f }
}

func OptMiddleware[T any](f func(T) (T, error)) func(*serviceImpl[T]) {
	return func(s *serviceImpl[T]) {
		init := s.init
		s.init = func() (T, error) {
			val, err := init()
			if err != nil {
				return empty[T](), err
			}

			return f(val)
		}
	}
}

func OptNoReuse[T any]() func(*serviceImpl[T]) {
	return func(s *serviceImpl[T]) { s.noReuse = true }
}

func OptDeinit[T any](f func(T) error) func(*serviceImpl[T]) {
	return func(s *serviceImpl[T]) { s.deinit = f }
}
