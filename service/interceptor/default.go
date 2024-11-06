package interceptor

import (
	"context"
	"github.com/awakari/source-sse/service/writer"
	"github.com/r3labs/sse/v2"
)

type defaultInterceptor struct {
	w writer.Service
}

func NewDefault(w writer.Service) Interceptor {
	return defaultInterceptor{
		w: w,
	}
}

func (d defaultInterceptor) Handle(ctx context.Context, url string, ssEvt *sse.Event, raw map[string]any) (matches bool, err error) {
	matches = true
	return
}
