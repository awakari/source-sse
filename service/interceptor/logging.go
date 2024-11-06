package interceptor

import (
	"context"
	"fmt"
	"github.com/r3labs/sse/v2"
	"log/slog"
)

type logging struct {
	i   Interceptor
	log *slog.Logger
	t   string
}

func NewLogging(i Interceptor, log *slog.Logger, t string) Interceptor {
	return logging{
		i:   i,
		log: log,
		t:   t,
	}
}

func (l logging) Handle(ctx context.Context, src string, ssEvt *sse.Event, raw map[string]any) (matches bool, err error) {
	if matches, err = l.i.Handle(ctx, src, ssEvt, raw); matches {
		switch err {
		case nil:
			l.log.Debug(fmt.Sprintf("interceptor(%s).Handle(%s, %s): ok", l.t, src, string(ssEvt.ID)))
		default:
			l.log.Error(fmt.Sprintf("interceptor(%s).Handle(%s, %s): %s", l.t, src, string(ssEvt.ID), err))
		}
	}
	return
}
