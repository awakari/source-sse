package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/source-sse/api/grpc"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service"
	"github.com/awakari/source-sse/service/handler"
	"github.com/awakari/source-sse/service/writer"
	"github.com/awakari/source-sse/storage/mongo"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

func main() {

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	log.Info("starting the update for the feeds")

	// determine the replica index
	replicaNameParts := strings.Split(cfg.Replica.Name, "-")
	if len(replicaNameParts) < 2 {
		panic("unable to parse the replica name: " + cfg.Replica.Name)
	}
	var replicaIndex int
	replicaIndex, err = strconv.Atoi(replicaNameParts[len(replicaNameParts)-1])
	if err != nil {
		panic(err)
	}
	if replicaIndex < 0 {
		panic(fmt.Sprintf("Negative replica index: %d", replicaIndex))
	}
	log.Info(fmt.Sprintf("Replica: %d", replicaIndex))

	var clientAwk api.Client
	clientAwk, err = api.
		NewClientBuilder().
		WriterUri(cfg.Api.Writer.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer clientAwk.Close()
	log.Info("initialized the Awakari API client")

	svcWriter := writer.NewService(clientAwk, cfg.Api.Writer.Backoff, cfg.Api.Writer.Cache, log)
	svcWriter = writer.NewLogging(svcWriter, log)

	ctx := context.Background()
	stor, err := mongo.NewStorage(ctx, cfg.Db)
	if err != nil {
		panic(err)
	}
	defer stor.Close()

	handlersLock := &sync.Mutex{}
	handlerByUrl := make(map[string]handler.Handler)
	svc := service.NewService(svcWriter, cfg.Api, cfg.Event, stor, uint32(replicaIndex), handlersLock, handlerByUrl, handler.New)
	svc = service.NewServiceLogging(svc, log)
	err = resumeHandlers(ctx, svc, svcWriter, uint32(replicaIndex), cfg, handlersLock, handlerByUrl)
	if err != nil {
		panic(err)
	}

	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	err = apiGrpc.Serve(cfg.Api.Port, svc)
	if err != nil {
		panic(err)
	}
}

func resumeHandlers(
	ctx context.Context,
	svc service.Service,
	svcWriter writer.Service,
	replicaIndex uint32,
	cfg config.Config,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
) (err error) {
	var cursor string
	var urls []string
	var str model.Stream
	for {
		urls, err = svc.List(ctx, 100, model.Filter{}, model.OrderAsc, cursor)
		if err == nil {
			if len(urls) == 0 {
				break
			}
			for _, url := range urls {
				str, err = svc.Read(ctx, url)
				if err == nil && str.Replica == replicaIndex {
					resumeHandler(ctx, url, str, svcWriter, cfg, handlersLock, handlerByUrl)
				}
				if err != nil {
					break
				}
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func resumeHandler(
	ctx context.Context,
	url string,
	str model.Stream,
	w writer.Service,
	cfg config.Config,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	h := handler.New(url, str, cfg.Api, cfg.Event, w)
	handlerByUrl[url] = h
	go h.Handle(ctx)
}
