package storage

import (
	"context"
	"errors"
	"github.com/awakari/source-sse/model"
	"io"
)

type Storage interface {
	io.Closer
	Create(ctx context.Context, url string, str model.Stream) (err error)
	Read(ctx context.Context, url string) (str model.Stream, err error)
	Delete(ctx context.Context, url, groupId, userId string) (err error)
	List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error)
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrUnexpected = errors.New("unexpected error")
