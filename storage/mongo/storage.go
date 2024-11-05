package mongo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/awakari/source-sse/config"
	"github.com/awakari/source-sse/model"
	"github.com/awakari/source-sse/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type storageMongo struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

type record struct {
	Url          string    `bson:"url"`
	Auth         string    `bson:"auth,omitempty"`
	GroupId      string    `bson:"gid"`
	UserId       string    `bson:"uid,omitempty"`
	ReplicaIndex uint32    `bson:"ridx"`
	CreatedAt    time.Time `bson:"createdAt"`
}

const attrUrl = "url"
const attrAuth = "auth"
const attrGroupId = "gid"
const attrUserId = "uid"
const attrReplicaIndex = "ridx"
const attrCreatedAt = "createdAt"

var optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
var optsGet = options.
	FindOne().
	SetShowRecordID(false).
	SetProjection(projRead)
var projRead = bson.D{
	{
		Key:   attrAuth,
		Value: 1,
	},
	{
		Key:   attrGroupId,
		Value: 1,
	},
	{
		Key:   attrUserId,
		Value: 1,
	},
	{
		Key:   attrReplicaIndex,
		Value: 1,
	},
	{
		Key:   attrCreatedAt,
		Value: 1,
	},
}
var projList = bson.D{
	{
		Key:   attrUrl,
		Value: 1,
	},
}
var sortListAsc = bson.D{
	{
		Key:   attrUrl,
		Value: 1,
	},
}
var sortListDesc = bson.D{
	{
		Key:   attrUrl,
		Value: -1,
	},
}

func NewStorage(ctx context.Context, cfgDb config.DbConfig) (s storage.Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
	if cfgDb.Tls.Enabled {
		clientOpts = clientOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cfgDb.Tls.Insecure})
	}
	if len(cfgDb.UserName) > 0 {
		auth := options.Credential{
			Username:    cfgDb.UserName,
			Password:    cfgDb.Password,
			PasswordSet: len(cfgDb.Password) > 0,
		}
		clientOpts = clientOpts.SetAuth(auth)
	}
	conn, err := mongo.Connect(ctx, clientOpts)
	var sm storageMongo
	if err == nil {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Name)
		sm.conn = conn
		sm.db = db
		sm.coll = coll
		_, err = sm.ensureIndices(ctx, cfgDb.Table.Retention)
	}
	if err == nil {
		s = sm
	}
	return
}

func (sm storageMongo) ensureIndices(ctx context.Context, retentionPeriod time.Duration) ([]string, error) {
	return sm.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{
					Key:   attrUrl,
					Value: 1,
				},
			},
			Options: options.
				Index().
				SetUnique(true),
		},
	})
}

func (sm storageMongo) Close() error {
	return sm.conn.Disconnect(context.TODO())
}

func (sm storageMongo) Create(ctx context.Context, url string, str model.Stream) (err error) {
	_, err = sm.coll.InsertOne(ctx, record{
		Url:          url,
		Auth:         str.Auth,
		GroupId:      str.GroupId,
		UserId:       str.UserId,
		ReplicaIndex: str.Replica,
		CreatedAt:    str.CreatedAt.UTC(),
	})
	err = decodeError(err, url)
	return
}

func (sm storageMongo) Read(ctx context.Context, url string) (str model.Stream, err error) {
	q := bson.M{
		attrUrl: url,
	}
	var result *mongo.SingleResult
	result = sm.coll.FindOne(ctx, q, optsGet)
	err = result.Err()
	var rec record
	if err == nil {
		err = result.Decode(&rec)
	}
	if err == nil {
		str.Auth = rec.Auth
		str.CreatedAt = rec.CreatedAt.UTC()
		str.GroupId = rec.GroupId
		str.UserId = rec.UserId
		str.Replica = rec.ReplicaIndex
	}
	err = decodeError(err, url)
	return
}

func (sm storageMongo) Delete(ctx context.Context, url, groupId, userId string) (err error) {
	var result *mongo.DeleteResult
	result, err = sm.coll.DeleteOne(ctx, bson.M{
		attrUrl:     url,
		attrGroupId: groupId,
		attrUserId:  userId,
	})
	switch err {
	case nil:
		if result.DeletedCount < 1 {
			err = fmt.Errorf("%w by url %s", storage.ErrNotFound, url)
		}
	default:
		err = decodeError(err, url)
	}
	return
}

func (sm storageMongo) List(ctx context.Context, limit uint32, filter model.Filter, order model.Order, cursor string) (urls []string, err error) {
	q := bson.M{}
	if filter.UserId != "" {
		q[attrGroupId] = filter.GroupId
		q[attrUserId] = filter.UserId
	}
	optsList := options.
		Find().
		SetLimit(int64(limit)).
		SetShowRecordID(false).
		SetProjection(projList)
	var clauseCursor bson.M
	switch order {
	case model.OrderDesc:
		clauseCursor = bson.M{
			"$lt": cursor,
		}
		optsList = optsList.SetSort(sortListDesc)
	default:
		clauseCursor = bson.M{
			"$gt": cursor,
		}
		optsList = optsList.SetSort(sortListAsc)
	}
	q["$and"] = []bson.M{
		{
			attrUrl: clauseCursor,
		},
		{
			"$or": []bson.M{
				{
					attrUrl: bson.M{
						"$regex": filter.Pattern,
					},
				},
				{
					attrUrl: bson.M{
						"$regex": filter.Pattern,
					},
				},
			},
		},
	}
	var cur *mongo.Cursor
	cur, err = sm.coll.Find(ctx, q, optsList)
	if err == nil {
		for cur.Next(ctx) {
			var rec record
			err = errors.Join(err, cur.Decode(&rec))
			if err == nil {
				urls = append(urls, rec.Url)
			}
		}
	}
	err = decodeError(err, cursor)
	return
}

func decodeError(src error, url string) (dst error) {
	switch {
	case src == nil:
	case errors.Is(src, mongo.ErrNoDocuments):
		dst = fmt.Errorf("%w: %s", storage.ErrNotFound, url)
	case mongo.IsDuplicateKeyError(src):
		dst = fmt.Errorf("%w: %s", storage.ErrConflict, url)
	default:
		dst = fmt.Errorf("%w: %s", storage.ErrUnexpected, src)
	}
	return
}
