package otel_zero

import (
	"context"
	"io"
	"net/http"

	"github.com/Himan000/otel_zap_logger/otel"
	"github.com/Himan000/otel_zap_logger/propagation/extract"
	"github.com/Himan000/zero_mdc_log"
	zero "github.com/Himan000/zero_mdc_log/log"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func Init(r *gin.Engine) {
	otel.Default.Init(r)
	zero.Init()
	r.Use(LogContextMiddleware)
}

// 调用链的请求包装
func NewReqeust(method, url string, body io.Reader) (*http.Response, error) {
	request, _ := http.NewRequest(method, url, body)
	ctx, _ := zero.MDC().Get("ctx")
	otel.HttpInject(ctx.(context.Context), request)
	client := &http.Client{}

	traceparent := request.Header.Get("Traceparent")
	b3 := request.Header.Get("B3")

	if traceparent == "" && b3 != "" {
		traceparent = extract.ConvertTraceIdFromB3ToTraceparentFormat(b3)
		request.Header.Add("traceparent", traceparent)
	}

	res, err := client.Do(request)
	return res, err
}

// 将日志相关需要协程安全的信息放到MDC
func LogContextMiddleware(c *gin.Context) {
	ctx := otel.Default.Start(c.Request.Context())
	zero.MDC().Set("ctx", ctx)                          // 将ctx房到MDC，后面request请求可以用
	zero.MDC().Set(zero.TRACE_ID, otel.GetTraceId(ctx)) // 这个是调用链的trace_id
	c.Next()
	otel.End()
}

func MDC() *zero_mdc_log.MdcAdapter {
	return zero.MDC()
}

func Log() *zero_mdc_log.Overlog {
	return zero.Log()
}

func Info() *zerolog.Event {
	return zero.Log().Info()
}

func Debug() *zerolog.Event {
	return zero.Log().Debug()
}

func Error() *zerolog.Event {
	return zero.Log().Error()
}

func Panic() *zerolog.Event {
	return zero.Log().Panic()
}

// 设置请求的middleware，每个请求写一些日志
func SetLogger(configItems ...zero_mdc_log.ConfigItem) gin.HandlerFunc {
	return zero_mdc_log.SetLogger(configItems...)
}
