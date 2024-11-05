# source-sse
Server-sent events source type

```shell
grpcurl \
  -plaintext \
  -proto api/grpc/service.proto \
  -d @ \
  localhost:50051 \
  awakari.source.sse.Service/Create
```

```json
{
  "url": "https://stream.wikimedia.org/v2/stream/recentchange",
  "groupId": "default"
}
```

```shell
grpcurl \
  -plaintext \
  -proto api/grpc/service.proto \
  -d @ \
  localhost:50051 \
  awakari.source.sse.Service/Delete
```

```json
{
  "url": "https://stream.wikimedia.org/v2/stream/recentchange",
  "groupId": "default"
}
```
