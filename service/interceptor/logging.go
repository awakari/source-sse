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

func (l logging) Matches(ssEvt *sse.Event, raw map[string]any) (matches bool) {
	matches = l.i.Matches(ssEvt, raw)
	if matches {
		l.log.Debug(fmt.Sprintf("interceptor(%s) matches event: %s, id: %s", l.t, string(ssEvt.Event), string(ssEvt.ID)))
	}
	return
}

func (l logging) Handle(ctx context.Context, ssEvt *sse.Event) (err error) {
	err = l.i.Handle(ctx, ssEvt)
	switch err {
	case nil:
		l.log.Debug(fmt.Sprintf("interceptor(%s).Handle(%s): ok", l.t, string(ssEvt.ID)))
	default:
		l.log.Error(fmt.Sprintf("interceptor(%s).Handle(%s): %s", l.t, string(ssEvt.ID), err))
	}
	return
}
