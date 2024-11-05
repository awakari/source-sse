package handler

import (
	"context"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service/writer"
)

type mockHandler struct{}

var NewMock Factory = func(url string, str model.Stream, cfgApi config.ApiConfig, cfgEvt config.EventConfig, w writer.Service) Handler {
	return mockHandler{}
}

func (m mockHandler) Close() error {
	return nil
}

func (m mockHandler) Handle(ctx context.Context) {
	return
}
