package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service/handler"
	"github.com/awakari/source-sse/storage"
	"sync"
	"time"
)

type Service interface {
	Create(ctx context.Context, url, auth, groupId, userId string, at time.Time) (err error)
	Read(ctx context.Context, url string) (str model.Stream, err error)
	Delete(ctx context.Context, url, groupId, userId string) (err error)
	List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error)
}

type svc struct {
	stor           storage.Storage
	replicaIndex   uint32
	handlersLock   *sync.Mutex
	handlerByUrl   map[string]handler.Handler
	handlerFactory handler.Factory
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrUnexpected = errors.New("unexpected")

func NewService(
	stor storage.Storage,
	replicaIndex uint32,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
	handlerFactory handler.Factory,
) Service {
	return svc{
		stor:           stor,
		replicaIndex:   replicaIndex,
		handlersLock:   handlersLock,
		handlerByUrl:   handlerByUrl,
		handlerFactory: handlerFactory,
	}
}

func (s svc) Create(ctx context.Context, url, auth, groupId, userId string, at time.Time) (err error) {
	str := model.Stream{
		Auth:      auth,
		GroupId:   groupId,
		UserId:    userId,
		CreatedAt: at,
		Replica:   s.replicaIndex,
	}
	err = s.stor.Create(ctx, url, str)
	if err == nil {
		s.handlersLock.Lock()
		defer s.handlersLock.Unlock()
		h := s.handlerFactory(url, str)
		s.handlerByUrl[url] = h
		go h.Handle(context.Background())
	}
	err = translateError(err)
	return
}

func (s svc) Read(ctx context.Context, url string) (str model.Stream, err error) {
	str, err = s.stor.Read(ctx, url)
	err = translateError(err)
	return
}

func (s svc) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	err = s.stor.Delete(ctx, url, groupId, userId)
	if err == nil {
		s.handlersLock.Lock()
		defer s.handlersLock.Unlock()
		h, hOk := s.handlerByUrl[url]
		if hOk {
			err = h.Close()
		}
	}
	err = translateError(err)
	return
}

func (s svc) List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error) {
	urls, err = s.stor.List(ctx, limit, filter, order, cursor)
	err = translateError(err)
	return
}

func translateError(src error) (dst error) {
	switch {
	case errors.Is(src, storage.ErrConflict):
		dst = fmt.Errorf("%w: %s", ErrConflict, src)
	case errors.Is(src, storage.ErrNotFound):
		dst = fmt.Errorf("%w: %s", ErrNotFound, src)
	case src != nil:
		dst = fmt.Errorf("%w: %s", ErrUnexpected, src)
	}
	return
}
