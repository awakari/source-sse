package interceptor

import (
	"context"
	"github.com/r3labs/sse/v2"
)

type Interceptor interface {
	Matches(et []byte, raw map[string]any) (matches bool)
	Handle(ctx context.Context, ssEvt *sse.Event) (err error)
}
