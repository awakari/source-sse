package grpc

import (
	"context"
	"errors"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type controller struct {
	svc service.Service
}

func NewController(svc service.Service) ServiceServer {
	return controller{
		svc: svc,
	}
}

func (c controller) Create(ctx context.Context, req *CreateRequest) (resp *CreateResponse, err error) {
	resp = &CreateResponse{}
	switch req.Url {
	case "":
		err = status.Error(codes.InvalidArgument, "empty url")
	default:
		err = c.svc.Create(ctx, req.Url, req.Auth, req.GroupId, req.UserId, time.Now().UTC())
		err = translateError(err)
	}
	return
}

func (c controller) Read(ctx context.Context, req *ReadRequest) (resp *ReadResponse, err error) {
	resp = &ReadResponse{}
	var str model.Stream
	str, err = c.svc.Read(ctx, req.Url)
	if err == nil {
		resp.GroupId = str.GroupId
		resp.UserId = str.UserId
		resp.CreatedAt = timestamppb.New(str.CreatedAt.UTC())
	}
	err = translateError(err)
	return
}

func (c controller) Delete(ctx context.Context, req *DeleteRequest) (resp *DeleteResponse, err error) {
	resp = &DeleteResponse{}
	err = c.svc.Delete(ctx, req.Url, req.GroupId, req.UserId)
	err = translateError(err)
	return
}

func (c controller) List(ctx context.Context, req *ListRequest) (resp *ListResponse, err error) {
	resp = &ListResponse{}
	filter := model.Filter{}
	if req.Filter != nil {
		filter.GroupId = req.Filter.GroupId
		filter.UserId = req.Filter.UserId
		filter.Pattern = req.Filter.Pattern
	}
	var order model.Order
	switch req.Order {
	case Order_DESC:
		order = model.OrderDesc
	}
	var urls []string
	urls, err = c.svc.List(ctx, req.Limit, filter, order, req.Cursor)
	if err == nil {
		resp.Urls = urls
	}
	err = translateError(err)
	return
}

func translateError(src error) (dst error) {
	switch {
	case errors.Is(src, service.ErrNotFound):
		dst = status.Error(codes.NotFound, src.Error())
	case errors.Is(src, service.ErrConflict):
		dst = status.Error(codes.AlreadyExists, src.Error())
	case src != nil:
		dst = status.Error(codes.Internal, src.Error())
	}
	return
}
