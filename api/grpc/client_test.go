package grpc

import (
	"context"
	"fmt"
	"github.com/awakari/source-sse/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
	"os"
	"testing"
	"time"
)

var port uint16 = 50051

var log = slog.Default()

func TestMain(m *testing.M) {
	svc := service.NewServiceMock()
	svc = service.NewServiceLogging(svc, log)
	go func() {
		err := Serve(port, svc)
		if err != nil {
			log.Error(err.Error())
		}
	}()
	code := m.Run()
	os.Exit(code)
}

func TestServiceClient_Create(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req *CreateRequest
		err error
	}{
		"ok": {
			req: &CreateRequest{
				Url: "url0",
			},
		},
		"empty url": {
			req: &CreateRequest{},
			err: status.Error(codes.InvalidArgument, "empty url"),
		},
		"fail": {
			req: &CreateRequest{
				Url: "fail",
			},
			err: status.Error(codes.Internal, "unexpected"),
		},
		"conflict": {
			req: &CreateRequest{
				Url: "conflict",
			},
			err: status.Error(codes.AlreadyExists, "conflict"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.Create(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceClient_Read(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req       *ReadRequest
		groupId   string
		userId    string
		createdAt *timestamppb.Timestamp
		err       error
	}{
		"ok": {
			req: &ReadRequest{
				Url: "url0",
			},
			groupId:   "group0",
			userId:    "user1",
			createdAt: timestamppb.New(time.Date(2024, 11, 4, 14, 52, 0, 0, time.UTC)),
		},
		"fail": {
			req: &ReadRequest{
				Url: "fail",
			},
			err: status.Error(codes.Internal, "unexpected"),
		},
		"missing": {
			req: &ReadRequest{
				Url: "missing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.Read(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.groupId, resp.GroupId)
				assert.Equal(t, c.userId, resp.UserId)
				assert.Equal(t, c.createdAt, resp.CreatedAt)
			}
		})
	}
}

func TestServiceClient_Delete(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req *DeleteRequest
		err error
	}{
		"ok": {
			req: &DeleteRequest{
				Url: "url0",
			},
		},
		"fail": {
			req: &DeleteRequest{
				Url: "fail",
			},
			err: status.Error(codes.Internal, "unexpected"),
		},
		"missing": {
			req: &DeleteRequest{
				Url: "missing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			_, err := client.Delete(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestServiceClient_List(t *testing.T) {
	//
	addr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	client := NewServiceClient(conn)
	//
	cases := map[string]struct {
		req  *ListRequest
		urls []string
		err  error
	}{
		"ok1": {
			req: &ListRequest{},
			urls: []string{
				"url0",
				"url1",
			},
		},
		"ok2": {
			req: &ListRequest{
				Filter: &Filter{
					GroupId: "group1",
					UserId:  "user2",
					Pattern: "foo",
				},
				Order: Order_DESC,
			},
			urls: []string{
				"url0",
				"url1",
			},
		},
		"fail": {
			req: &ListRequest{
				Cursor: "fail",
			},
			err: status.Error(codes.Internal, "unexpected"),
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			resp, err := client.List(context.TODO(), c.req)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.urls, resp.Urls)
			}
		})
	}
}
