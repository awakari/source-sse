package service

import (
	"context"
	"fmt"
	"github.com/awakari/source-sse/model"
	"log/slog"
	"time"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewServiceLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) Create(ctx context.Context, url, auth, groupId, userId string, at time.Time) (err error) {
	err = l.svc.Create(ctx, url, auth, groupId, userId, at)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.Create(%s, %s, %s, %s, %s): %s", url, auth, groupId, userId, at, err))
	return
}

func (l logging) Read(ctx context.Context, url string) (str model.Stream, err error) {
	str, err = l.svc.Read(ctx, url)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.Read(%s): %+v, %s", url, str, err))
	return
}

func (l logging) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	err = l.svc.Delete(ctx, url, groupId, userId)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.Delete(%s, %s/%s): %s", url, groupId, userId, err))
	return
}

func (l logging) List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error) {
	urls, err = l.svc.List(ctx, limit, filter, order, cursor)
	l.log.Log(context.TODO(), logLevel(err), fmt.Sprintf("service.List(%d, %+v, %+v, %s): %d, %s", limit, filter, order, cursor, len(urls), err))
	return
}

func logLevel(err error) (lvl slog.Level) {
	switch err {
	case nil:
		lvl = slog.LevelDebug
	default:
		lvl = slog.LevelError
	}
	return
}
