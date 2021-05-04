package common

import (
	"github.com/alessiosavi/GoGPUtils/helper"
	"github.com/valyala/fasthttp"
)

type req struct {
	Method    string `json:"method"`
	Url       string `json:"url"`
	UserAgent string `json:"user_agent"`
	Headers   string `json:"headers"`
	Body      string `json:"body"`
}

func DebugRequest(ctx *fasthttp.RequestCtx) string {
	var r req
	r.Url = ctx.URI().String()
	r.Method = string(ctx.Method())
	r.UserAgent = string(ctx.UserAgent())
	r.Headers = ctx.Request.Header.String()
	r.Body = string(ctx.Request.Body())
	return helper.Marshal(r)
}
