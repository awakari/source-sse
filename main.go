package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/source-sse/api/grpc"
	"github.com/awakari/source-sse/api/grpc/events"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/service"
	"github.com/awakari/source-sse/service/handler"
	"github.com/awakari/source-sse/service/interceptor"
	"github.com/awakari/source-sse/service/writer"
	"github.com/awakari/source-sse/storage/mongo"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	log.Info(fmt.Sprintf("Replica: %d", replicaIndex))
	if replicaIndex < 1 {
		log.Info("Replica is dummy, sleeping forever...")
		select {}
	}

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

	connPoolEvts, err := grpcpool.New(
		func() (*grpc.ClientConn, error) {
			return grpc.NewClient(cfg.Api.Events.Uri, grpc.WithTransportCredentials(insecure.NewCredentials()))
		},
		int(cfg.Api.Events.Connection.Count.Init),
		int(cfg.Api.Events.Connection.Count.Max),
		cfg.Api.Events.Connection.IdleTimeout,
	)
	if err != nil {
		panic(err)
	}
	defer connPoolEvts.Close()
	clientEvts := events.NewClientPool(connPoolEvts)
	svcEvts := events.NewService(clientEvts)
	svcEvts = events.NewLoggingMiddleware(svcEvts, log)
	err = svcEvts.SetStream(context.TODO(), cfg.Api.Events.Topics.Mastodon, cfg.Api.Events.Limit)
	if err != nil {
		panic(err)
	}

	var pubMastodon events.Writer
	pubMastodon, err = svcEvts.NewPublisher(ctx, cfg.Api.Events.Topics.Mastodon)
	if err != nil {
		panic(err)
	}
	defer pubMastodon.Close()

	interceptors := []interceptor.Interceptor{
		interceptor.NewLogging(interceptor.NewMastodon(cfg.Api.Events, pubMastodon), log, "mastodon"),
		interceptor.NewLogging(interceptor.NewWikiMedia(svcWriter, cfg.Api.GroupId, cfg.Event.Type), log, "wikimedia"),
		interceptor.NewLogging(interceptor.NewDefault(svcWriter), log, "default"),
	}

	handlersLock := &sync.Mutex{}
	handlerByUrl := make(map[string]handler.Handler)
	handlerFactory := handler.NewFactory(cfg.Api, cfg.Event, interceptors)

	svc := service.NewService(stor, uint32(replicaIndex), handlersLock, handlerByUrl, handlerFactory)
	svc = service.NewServiceLogging(svc, log)
	err = resumeHandlers(ctx, log, svc, uint32(replicaIndex), handlersLock, handlerByUrl, handlerFactory)
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
	log *slog.Logger,
	svc service.Service,
	replicaIndex uint32,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
	handlerFactory handler.Factory,
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
			cursor = urls[len(urls)-1]
			for _, url := range urls {
				str, err = svc.Read(ctx, url)
				if err == nil && str.Replica == replicaIndex {
					resumeHandler(ctx, log, url, str, handlersLock, handlerByUrl, handlerFactory)
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
	log *slog.Logger,
	url string,
	str model.Stream,
	handlersLock *sync.Mutex,
	handlerByUrl map[string]handler.Handler,
	handlerFactory handler.Factory,
) {
	handlersLock.Lock()
	defer handlersLock.Unlock()
	h := handlerFactory(url, str)
	handlerByUrl[url] = h
	go h.Handle(ctx)
	log.Info(fmt.Sprintf("resumed handler for %s", url))
}
