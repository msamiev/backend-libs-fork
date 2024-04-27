package validator

import "github.com/labstack/echo/v4"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func reportErr(c echo.Context, code int, msg string) error {
	return c.JSON(code, Error{Code: code, Message: msg})
}
