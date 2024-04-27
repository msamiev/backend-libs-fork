package di

import (
	"errors"
	"strconv"
	"testing"
)

func TestReleaseErrfmt(t *testing.T) {
	c := New()

	err := c.Release()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	c.errors = append(c.errors, errors.New("1"))
	err = c.Release()

	if err.Error() != "1" {
		t.Errorf("Unexpected error: %s", err)
	}

	c.errors = append(c.errors, errors.New("2"))
	err = c.Release()

	if err.Error() != "1: 2" {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestInitOnlyNeeded(_ *testing.T) {
	c := New()

	Set(c, OptInit(func() (map[string]any, error) {
		return map[string]any{"a": 1}, nil
	}))
	Set(c, OptInit(func() ([]int, error) {
		return []int{2}, nil
	}))
	Set(c, OptInit(func() (string, error) {
		panic("no need to init 'string'")
	}))
	SetNamed(c, "main", OptInit(func() (any, error) {
		return map[string]any{
			"b": Get[map[string]any](c),
			"c": Get[[]int](c),
		}, nil
	}))

	GetNamed[any](c, "main")
}

func TestReuse(t *testing.T) {
	c := New()
	count := new(int)

	Set(c, OptInit(func() (*int, error) {
		*count++
		return count, nil
	}), OptNoReuse[*int]())

	for i := 1; i < 5; i++ {
		if val := Get[*int](c); *val != i {
			t.Errorf("Unexpected val: %v", *val)
		}
	}
}

func TestReuseWithMiddleware(t *testing.T) {
	c := New()
	count := new(int)

	Set(c, OptInit(func() (*int, error) {
		*count++
		return count, nil
	}), OptMiddleware(func(i *int) (*int, error) {
		*i--
		return i, nil
	}), OptNoReuse[*int]())

	for i := 0; i < 5; i++ {
		if val := Get[*int](c); *val != 0 {
			t.Errorf("Unexpected val: %v", *val)
		}
	}
}

func TestDeinit(t *testing.T) {
	var (
		c    = New()
		err1 = errors.New("1")
		err2 = errors.New("2")
		err3 = errors.New("3")
	)

	Set(c, OptInit(func() (int, error) {
		return 42, nil
	}), OptDeinit(func(int) error {
		return err1
	}))
	Set(c, OptInit(func() (string, error) {
		return strconv.Itoa(Get[int](c)), nil
	}), OptDeinit(func(string) error {
		return err2
	}))
	SetNamed(c, "format", OptInit(func() (string, error) {
		return "format: " + Get[string](c), nil
	}), OptDeinit(func(string) error {
		return err3
	}))

	result := GetNamed[string](c, "format")
	if result != "format: 42" {
		t.Errorf("Unexpected: %v", result)
	}

	err := c.Release()
	if err == nil {
		t.Errorf("Release should return error")
	}

	if err.Error() != "3: 2: 1" {
		t.Errorf("Unexpected: %v", err)
	}
}

func TestMultiDeinit(t *testing.T) {
	var (
		c    = New()
		err1 = errors.New("1")
	)

	Set(c, OptInit(func() (int, error) {
		return 42, nil
	}), OptDeinit(func(int) error {
		return err1
	}))

	one := Get[int](c)
	if one != 42 {
		t.Errorf("Unexpected: %d", one)
	}

	two := Get[int](c)
	if two != 42 {
		t.Errorf("Unexpected: %d", two)
	}

	err := c.Release()
	if err == nil {
		t.Errorf("Release should return error")
	}

	if err.Error() != "1" {
		t.Errorf("Unexpected: %v", err)
	}
}
