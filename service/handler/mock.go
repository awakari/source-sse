package handler

import (
	"context"
	"github.com/awakari/source-sse/model"
)

type mockHandler struct{}

var NewMock Factory = func(url string, str model.Stream) Handler {
	return mockHandler{}
}

func (m mockHandler) Close() error {
	return nil
}

func (m mockHandler) Handle(ctx context.Context) {
	return
}
