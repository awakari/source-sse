package interceptor

import (
	"context"
	"github.com/awakari/source-sse/api/grpc/events"
	"github.com/awakari/source-sse/config"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/r3labs/sse/v2"
	"github.com/segmentio/ksuid"
)

type mastodon struct {
	cfgEvts config.EventsConfig
	w       events.Writer
}

const ceSpecVersion = "1.0"

func NewMastodon(cfgEvts config.EventsConfig, w events.Writer) Interceptor {
	return mastodon{
		cfgEvts: cfgEvts,
		w:       w,
	}
}

func (m mastodon) Matches(et []byte, raw map[string]any) (matches bool) {
	if _, accOk := raw["account"]; !accOk {
		return false
	}
	if _, visibilityOk := raw["visibility"]; !visibilityOk {
		return false
	}
	if _, contentOk := raw["content"]; !contentOk {
		return false
	}
	if _, uriOk := raw["uri"]; !uriOk {
		return false
	}
	if _, idOk := raw["id"]; !idOk {
		return false
	}
	return true
}

func (m mastodon) Handle(ctx context.Context, ssEvt *sse.Event) (err error) {
	evt := &pb.CloudEvent{
		Id:          ksuid.New().String(),
		Source:      m.cfgEvts.Source,
		SpecVersion: ceSpecVersion,
		Type:        string(ssEvt.Event),
		Data: &pb.CloudEvent_BinaryData{
			BinaryData: ssEvt.Data,
		},
	}
	var ackCount uint32
	ackCount, err = m.w.Write(ctx, []*pb.CloudEvent{
		evt,
	})
	if err != nil {
		panic("mastodon interceptor: failed to write event: " + err.Error())
	}
	if ackCount < 1 {
		panic("mastodon interceptor: failed to acknowledge event")
	}
	return
}
