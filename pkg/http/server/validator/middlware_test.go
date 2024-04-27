package validator_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/matryer/is"

	"github.com/fusionmedialimited/backend-infra-libraries/v3/pkg/http/server/validator"
)

const (
	host = "http://127.0.0.1"
)

func newServer(isLib *is.I) *echo.Echo {
	var (
		e = echo.New()
	)

	spec, err := (&openapi3.Loader{ReadFromURIFunc: openapi3.ReadFromFile}).LoadFromFile("test-spec.yaml")
	isLib.NoErr(err)
	e.Use(validator.NewMiddlewareFunc(spec))

	e.GET("/v1/watchlist", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []echo.Map{
			{
				"id":             1,
				"name":           "My Test Watchlist",
				"currency_id":    1,
				"source":         "desktop",
				"instrument_ids": []int64{1},
			},
		})
	})

	e.POST("/v1/watchlist", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []echo.Map{
			{
				"id": 1,
			},
		})
	})

	e.DELETE("/v1/watchlist/:watchlistID/instrument/:instrumentID", func(ctx echo.Context) error {
		watchlistID := ctx.Param("watchlistID")
		instrumentID := ctx.Param("instrumentID")
		return ctx.JSON(http.StatusOK, echo.Map{
			"watchlist_id":   watchlistID,
			"instruments_id": instrumentID,
		})
	})
	return e
}

func TestServer(t *testing.T) {
	var (
		isLib = is.New(t)
		e     = newServer(isLib)
	)
	req := httptest.NewRequest(http.MethodGet, host+"/v1/watchlist?a=a&b=b&c=c", nil)
	req.Header.Set("X-User-Id", "1")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	isLib.Equal(http.StatusOK, res.Code)
}

func TestBadPath(t *testing.T) {
	var (
		isLib = is.New(t)
		e     = newServer(isLib)
	)
	req := httptest.NewRequest(http.MethodGet, host+"/v1/portfolio", nil)
	req.Header.Set("X-User-Id", "1")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	isLib.Equal(http.StatusNotFound, res.Code)
}

func TestBadMethod(t *testing.T) {
	var (
		isLib = is.New(t)
		e     = newServer(isLib)
	)
	req := httptest.NewRequest(http.MethodPut, host+"/v1/watchlist", nil)
	req.Header.Set("X-User-Id", "1")
	res := httptest.NewRecorder()
	e.ServeHTTP(res, req)
	isLib.Equal(http.StatusMethodNotAllowed, res.Code)
}

func TestParams(t *testing.T) {
	var (
		isLib = is.New(t)
		e     = newServer(isLib)
	)
	for _, tc := range []struct {
		testName     string
		target       string
		expectedCode int
	}{
		{
			testName:     "Correct Params",
			target:       host + "/v1/watchlist/1/instrument/1",
			expectedCode: http.StatusOK,
		},
		{
			testName:     "Bad Params",
			target:       host + "/v1/watchlist/something/instrument/something",
			expectedCode: http.StatusBadRequest,
		},
		{
			testName:     "Bad Params Schema",
			target:       host + "/v1/watchlist/{watchlistID}/instrument/{instrumentID}",
			expectedCode: http.StatusBadRequest,
		},
	} {
		tc := tc // capture range variable
		t.Run(tc.testName, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, tc.target, nil)
			req.Header.Set("X-User-Id", "1")
			res := httptest.NewRecorder()
			e.ServeHTTP(res, req)
			isLib.Equal(tc.expectedCode, res.Code)
		})
	}
}

func TestBody(t *testing.T) {
	var (
		isLib = is.New(t)
		e     = newServer(isLib)
	)
	for _, tc := range []struct {
		testName     string
		data         map[string]any
		expectedCode int
	}{
		{
			testName: "Correct Body",
			data: map[string]interface{}{
				"name":           "My Test Watchlist",
				"currency_id":    1,
				"source":         "android",
				"instrument_ids": []int64{1},
			},
			expectedCode: http.StatusOK,
		},
		{
			testName:     "Empty Body",
			data:         map[string]interface{}{},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName: "Wrong Source Enum Body",
			data: map[string]interface{}{
				"name":           "My Test Watchlist",
				"currency_id":    1,
				"source":         "xxxxxxxx",
				"instrument_ids": []int64{1},
			},
			expectedCode: http.StatusBadRequest,
		},
	} {
		tc := tc // capture range variable
		t.Run(tc.testName, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.data)
			isLib.NoErr(err)
			body := bytes.NewReader(jsonData)
			req := httptest.NewRequest(http.MethodPost, host+"/v1/watchlist", body)
			req.Header.Set("X-User-Id", "1")
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			e.ServeHTTP(res, req)
			isLib.Equal(tc.expectedCode, res.Code)
		})
	}
}
