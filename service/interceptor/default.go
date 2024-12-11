package interceptor

import (
	"context"
	"github.com/awakari/source-sse/api/http/pub"
	"github.com/r3labs/sse/v2"
)

type defaultInterceptor struct {
	svcPub pub.Service
}

func NewDefault(svcPub pub.Service) Interceptor {
	return defaultInterceptor{
		svcPub: svcPub,
	}
}

func (d defaultInterceptor) Handle(ctx context.Context, url string, ssEvt *sse.Event, raw map[string]any) (matches bool, err error) {
	matches = true
	return
}
