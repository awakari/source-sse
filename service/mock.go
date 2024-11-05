package service

import (
	"context"
	"github.com/awakari/source-sse/model"
	"time"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) Create(ctx context.Context, url, auth, groupId, userId string, at time.Time) (err error) {
	switch url {
	case "fail":
		err = ErrUnexpected
	case "conflict":
		err = ErrConflict
	}
	return
}

func (m mock) Read(ctx context.Context, url string) (str model.Stream, err error) {
	switch url {
	case "missing":
		err = ErrNotFound
	case "fail":
		err = ErrUnexpected
	default:
		str.GroupId = "group0"
		str.UserId = "user1"
		str.CreatedAt = time.Date(2024, 11, 4, 14, 52, 0, 0, time.UTC)
		str.Replica = 1
	}
	return
}

func (m mock) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	switch url {
	case "missing":
		err = ErrNotFound
	case "fail":
		err = ErrUnexpected
	}
	return
}

func (m mock) List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error) {
	switch cursor {
	case "fail":
		err = ErrUnexpected
	default:
		urls = []string{
			"url0",
			"url1",
		}
	}
	return
}
