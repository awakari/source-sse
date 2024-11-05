package handler

import (
	"context"
	"fmt"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service/interceptor"
	"github.com/bytedance/sonic"
	"github.com/r3labs/sse/v2"
	"io"
	"time"
)

type Handler interface {
	io.Closer
	Handle(ctx context.Context)
}

type handler struct {
	url          string
	str          model.Stream
	cfgApi       config.ApiConfig
	cfgEvt       config.SseConfig
	interceptors []interceptor.Interceptor

	clientSse *sse.Client
	chSsEvts  chan *sse.Event
}

type Factory func(url string, str model.Stream) Handler

func NewFactory(cfgApi config.ApiConfig, cfgEvt config.SseConfig, interceptors []interceptor.Interceptor) Factory {
	return func(url string, str model.Stream) Handler {
		return &handler{
			url:          url,
			str:          str,
			cfgApi:       cfgApi,
			cfgEvt:       cfgEvt,
			interceptors: interceptors,
		}
	}
}

func (h *handler) Close() error {
	h.clientSse.Unsubscribe(h.chSsEvts)
	return nil
}

func (h *handler) Handle(ctx context.Context) {
	var err error
	for {
		err = h.handleStream(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func (h *handler) handleStream(ctx context.Context) (err error) {
	h.clientSse = sse.NewClient(h.url)
	if h.str.Auth != "" {
		h.clientSse.Headers["Authorization"] = h.str.Auth
	}
	h.clientSse.Headers["User-Agent"] = h.cfgApi.UserAgent
	h.chSsEvts = make(chan *sse.Event)
	err = h.clientSse.SubscribeChanWithContext(ctx, "", h.chSsEvts)
	if err == nil {
		defer h.clientSse.Unsubscribe(h.chSsEvts)
		for {
			select {
			case ssEvt := <-h.chSsEvts:
				_ = h.handleStreamEvent(ctx, ssEvt)
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(h.cfgEvt.StreamTimeout):
				err = fmt.Errorf("timeout while wating for a new event from: %s", h.url)
			}
			if err != nil {
				break
			}
		}
	}
	return
}

func (h *handler) handleStreamEvent(ctx context.Context, ssEvt *sse.Event) (err error) {
	var raw map[string]any
	err = sonic.Unmarshal(ssEvt.Data, &raw)
	if err == nil {
		for _, i := range h.interceptors {
			if i.Matches(ssEvt.Event, raw) {
				err = i.Handle(ctx, ssEvt)
				break
			}
		}
	}
	return
}
