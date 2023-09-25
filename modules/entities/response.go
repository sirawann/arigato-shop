package entities

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirawann/arigato-shop/pkg/arigatologger"
)

type IResponse interface {
	Success(code int, data any) IResponse
	Error(code int, traceId, msg string) IResponse
	Res() error
}

type Response struct {
	StatusCode int
	Data       any
	ErrorRes   *ErrorResponse
	Context    *fiber.Ctx
	IsError    bool
}

type ErrorResponse struct {
	TraceId string `json:"traceId"`
	Msg     string `json:"message"`
}

func NewResponse(c *fiber.Ctx) IResponse {
	return &Response{
		Context: c,
	}
}

func (r *Response) Success(code int, data any) IResponse {
	r.StatusCode = code
	r.Data = data
	arigatologger.InitArigatoLogger(r.Context, &r.Data).Print().Save()
	return r
}
func (r *Response) Error(code int, traceId, msg string) IResponse {
	r.StatusCode = code
	r.ErrorRes = &ErrorResponse{
		TraceId: traceId,
		Msg:     msg,
	}
	r.IsError = true
	arigatologger.InitArigatoLogger(r.Context, &r.ErrorRes).Print().Save()
	return r
}

func (r *Response) Res() error {

	return r.Context.Status(r.StatusCode).JSON(func() any {
		if r.IsError {
			return &r.ErrorRes
		}
		return &r.Data
	}())

}

type PaginateRes struct {
	Data      any `json:"data"`
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	TotalPage int `json:"total_page"`
	TotalItem int `json:"total_item"`
}
