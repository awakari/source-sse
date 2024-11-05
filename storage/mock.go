package storage

import (
	"context"
	"github.com/awakari/source-sse/model"
	"time"
)

type mockStorage struct{}

func NewMockStorage() Storage {
	return mockStorage{}
}

func (m mockStorage) Close() error {
	//TODO implement me
	panic("implement me")
}

func (m mockStorage) Create(ctx context.Context, url string, str model.Stream) (err error) {
	switch url {
	case "fail":
		err = ErrUnexpected
	case "conflict":
		err = ErrConflict
	}
	return
}

func (m mockStorage) Read(ctx context.Context, url string) (str model.Stream, err error) {
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

func (m mockStorage) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	switch url {
	case "missing":
		err = ErrNotFound
	case "fail":
		err = ErrUnexpected
	}
	return
}

func (m mockStorage) List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error) {
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
