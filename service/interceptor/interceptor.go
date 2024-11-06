package interceptor

import (
	"context"
	"github.com/r3labs/sse/v2"
)

type Interceptor interface {
	Handle(ctx context.Context, src string, ssEvt *sse.Event, raw map[string]any) (matches bool, err error)
}
