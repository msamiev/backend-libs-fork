package validator

import (
	"net/http"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/labstack/echo/v4"
)

func NewMiddlewareFunc(spec *openapi3.T) echo.MiddlewareFunc {
	pathItemsMap := collectPathItemsMap(spec)
	return func(nextHandler echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			pathItem := pathItemsMap[c.Path()]
			if pathItem == nil {
				return nextHandler(c)
			}

			operation, ok := pathItem.Operations()[req.Method]
			if !ok {
				return nextHandler(c)
			}

			err := openapi3filter.ValidateRequest(
				req.Context(),
				&openapi3filter.RequestValidationInput{
					Request:     req,
					PathParams:  pathParams(c),
					QueryParams: c.QueryParams(),
					Route: &routers.Route{
						Spec:      spec,
						Path:      req.URL.Path,
						PathItem:  pathItem,
						Method:    req.Method,
						Operation: operation,
					},
					Options: &openapi3filter.Options{
						AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
					},
				})
			if err != nil {
				switch validateErr := err.(type) { //nolint:errorlint //it's ok
				case *openapi3filter.RequestError:
					return reportErr(c, http.StatusBadRequest, validateErr.Error())
				case *openapi3filter.SecurityRequirementsError:
					return reportErr(c, http.StatusForbidden, validateErr.Error())
				default:
					return reportErr(c, http.StatusInternalServerError, err.Error())
				}
			}
			return nextHandler(c)
		}
	}
}

func collectPathItemsMap(spec *openapi3.T) map[string]*openapi3.PathItem {
	pathItems := make(map[string]*openapi3.PathItem)
	for path, pathItem := range spec.Paths.Map() {
		if strings.Contains(path, "{") && strings.Contains(path, "}") {
			pathParts := strings.Split(path, "/")
			paramNames := make([]string, 0)
			for _, step := range pathParts {
				if strings.HasPrefix(step, "{") && strings.HasSuffix(step, "}") {
					paramNames = append(paramNames, step[1:len(step)-1])
				}
			}

			if len(paramNames) > 1 {
				sort.Slice(paramNames, func(i, j int) bool {
					return len(paramNames[i]) > len(paramNames[j])
				})
			}

			for _, name := range paramNames {
				path = strings.Replace(path, "/{"+name+"}", "/:"+name, 1)
			}
		}
		pathItems[path] = pathItem
	}
	return pathItems
}

func pathParams(c echo.Context) map[string]string {
	var (
		paramNames  = c.ParamNames()
		paramValues = c.ParamValues()
		res         = make(map[string]string)
	)
	for i, val := range paramValues {
		name := paramNames[i]
		res[name] = val
	}
	return res
}
