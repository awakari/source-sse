package service

import (
	"context"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service/handler"
	"github.com/awakari/source-sse/storage"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestService_Create(t *testing.T) {
	handlerByUrl := make(map[string]handler.Handler)
	s := NewService(storage.NewMockStorage(), 1, &sync.Mutex{}, handlerByUrl, handler.NewMock)
	s = NewServiceLogging(s, slog.Default())
	cases := map[string]struct {
		url          string
		auth         string
		groupId      string
		userId       string
		at           time.Time
		handlerCount int
		err          error
	}{
		"ok": {
			handlerCount: 1,
		},
		"fail": {
			url: "fail",
			err: ErrUnexpected,
		},
		"conflict": {
			url: "conflict",
			err: ErrConflict,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := s.Create(context.TODO(), c.url, c.auth, c.groupId, c.userId, c.at)
			assert.ErrorIs(t, err, c.err)
			assert.Equal(t, c.handlerCount, len(handlerByUrl))
			clear(handlerByUrl)
		})
	}
}

func TestService_Read(t *testing.T) {
	s := NewService(storage.NewMockStorage(), 1, &sync.Mutex{}, make(map[string]handler.Handler), handler.NewMock)
	s = NewServiceLogging(s, slog.Default())
	cases := map[string]struct {
		url string
		str model.Stream
		err error
	}{
		"ok": {
			str: model.Stream{
				GroupId:   "group0",
				UserId:    "user1",
				CreatedAt: time.Date(2024, 11, 4, 14, 52, 0, 0, time.UTC),
				Replica:   1,
			},
		},
		"fail": {
			url: "fail",
			err: ErrUnexpected,
		},
		"missing": {
			url: "missing",
			err: ErrNotFound,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			str, err := s.Read(context.TODO(), c.url)
			assert.ErrorIs(t, err, c.err)
			assert.Equal(t, c.str, str)
		})
	}
}

func TestService_Delete(t *testing.T) {
	s := NewService(storage.NewMockStorage(), 1, &sync.Mutex{}, make(map[string]handler.Handler), handler.NewMock)
	s = NewServiceLogging(s, slog.Default())
	cases := map[string]struct {
		url     string
		groupId string
		userId  string
		err     error
	}{
		"ok": {},
		"fail": {
			url: "fail",
			err: ErrUnexpected,
		},
		"missing": {
			url: "missing",
			err: ErrNotFound,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err := s.Delete(context.TODO(), c.url, c.groupId, c.userId)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestService_List(t *testing.T) {
	s := NewService(storage.NewMockStorage(), 1, &sync.Mutex{}, make(map[string]handler.Handler), handler.NewMock)
	s = NewServiceLogging(s, slog.Default())
	cases := map[string]struct {
		limit  uint32
		filter model.Filter
		order  model.Order
		cursor string
		urls   []string
		err    error
	}{
		"ok": {
			urls: []string{
				"url0",
				"url1",
			},
		},
		"fail": {
			cursor: "fail",
			err:    ErrUnexpected,
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			urls, err := s.List(context.TODO(), c.limit, c.filter, c.order, c.cursor)
			assert.ErrorIs(t, err, c.err)
			assert.Equal(t, c.urls, urls)
		})
	}
}
