package mongo

import (
	"context"
	"fmt"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var dbUri = os.Getenv("DB_URI_TEST_MONGO")

func TestNewStorage(t *testing.T) {
	//
	collName := fmt.Sprintf("tgchans-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Name = collName
	dbCfg.Table.Shard = false
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	//
	clear(ctx, t, s.(storageMongo))
}

func clear(ctx context.Context, t *testing.T, s storageMongo) {
	require.Nil(t, s.coll.Drop(ctx))
	require.Nil(t, s.Close())
}

func TestStorageMongo_Create(t *testing.T) {
	//
	collName := fmt.Sprintf("sse-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	sm := s.(storageMongo)
	defer clear(ctx, t, s.(storageMongo))
	//
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url0",
		Auth:         "token1",
		GroupId:      "group2",
		UserId:       "user3",
		ReplicaIndex: 4,
		CreatedAt:    time.Now().UTC(),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		url          string
		auth         string
		groupId      string
		userId       string
		replicaIndex uint32
		at           time.Time
		err          error
	}{
		"ok empty": {},
		"ok": {
			url:          "url1",
			auth:         "token1",
			groupId:      "group2",
			userId:       "user3",
			replicaIndex: 4,
			at:           time.Now().UTC(),
		},
		"dup id": {
			url: "url0",
			err: storage.ErrConflict,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = s.Create(ctx, c.url, model.Stream{
				Auth:      c.auth,
				GroupId:   c.groupId,
				UserId:    c.userId,
				CreatedAt: c.at,
				Replica:   c.replicaIndex,
			})
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_Read(t *testing.T) {
	//
	collName := fmt.Sprintf("sse-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	sm := s.(storageMongo)
	defer clear(ctx, t, s.(storageMongo))
	//
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url0",
		Auth:         "token1",
		GroupId:      "group2",
		UserId:       "user3",
		ReplicaIndex: 4,
		CreatedAt:    time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		url string
		out model.Stream
		err error
	}{
		"ok": {
			url: "url0",
			out: model.Stream{
				Auth:      "token1",
				GroupId:   "group2",
				UserId:    "user3",
				CreatedAt: time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
				Replica:   4,
			},
		},
		"missing": {
			url: "url1",
			err: storage.ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			var str model.Stream
			str, err = s.Read(ctx, c.url)
			assert.Equal(t, c.out, str)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_Delete(t *testing.T) {
	//
	collName := fmt.Sprintf("sse-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	sm := s.(storageMongo)
	defer clear(ctx, t, s.(storageMongo))
	//
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url0",
		Auth:         "token1",
		GroupId:      "group2",
		UserId:       "user3",
		ReplicaIndex: 4,
		CreatedAt:    time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
	})
	require.Nil(t, err)
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url1",
		Auth:         "token2",
		GroupId:      "group3",
		UserId:       "user4",
		ReplicaIndex: 5,
		CreatedAt:    time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		url     string
		groupId string
		userId  string
		err     error
	}{
		"ok": {
			url:     "url0",
			groupId: "group2",
			userId:  "user3",
		},
		"invalid url": {
			url:     "url2",
			groupId: "group3",
			userId:  "user4",
			err:     storage.ErrNotFound,
		},
		"missing user id": {
			url:     "url1",
			groupId: "group3",
			err:     storage.ErrNotFound,
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			err = s.Delete(ctx, c.url, c.groupId, c.userId)
			assert.ErrorIs(t, err, c.err)
		})
	}
}

func TestStorageMongo_List(t *testing.T) {
	//
	collName := fmt.Sprintf("sse-test-%d", time.Now().UnixMicro())
	dbCfg := config.DbConfig{
		Uri:  dbUri,
		Name: "sources",
	}
	dbCfg.Table.Name = collName
	dbCfg.Tls.Enabled = true
	dbCfg.Tls.Insecure = true
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()
	s, err := NewStorage(ctx, dbCfg)
	require.Nil(t, err)
	assert.NotNil(t, s)
	//
	sm := s.(storageMongo)
	defer clear(ctx, t, s.(storageMongo))
	//
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url0",
		Auth:         "token1",
		GroupId:      "group2",
		UserId:       "user3",
		ReplicaIndex: 4,
		CreatedAt:    time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
	})
	require.Nil(t, err)
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          "url1",
		Auth:         "token2",
		GroupId:      "group3",
		UserId:       "user4",
		ReplicaIndex: 5,
		CreatedAt:    time.Date(2024, 11, 4, 18, 49, 25, 0, time.UTC),
	})
	require.Nil(t, err)
	//
	cases := map[string]struct {
		limit  uint32
		filter model.Filter
		order  model.Order
		cursor string
		urls   []string
		err    error
	}{
		"asc": {
			urls: []string{
				"url0",
				"url1",
			},
		},
		"desc w/ limit": {
			cursor: "zzzz",
			limit:  1,
			order:  model.OrderDesc,
			urls: []string{
				"url1",
			},
		},
		"filter": {
			filter: model.Filter{
				GroupId: "group2",
				UserId:  "user3",
			},
			urls: []string{
				"url0",
			},
		},
		"cursor": {
			cursor: "url0",
			urls: []string{
				"url1",
			},
		},
	}
	//
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			urls, err := s.List(ctx, c.limit, c.filter, c.order, c.cursor)
			assert.ErrorIs(t, err, c.err)
			assert.Equal(t, c.urls, urls)
		})
	}
}
