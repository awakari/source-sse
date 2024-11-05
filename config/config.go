package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api   ApiConfig
	Db    DbConfig
	Event SseConfig
	Log   struct {
		Level int `envconfig:"LOG_LEVEL" default:"-4" required:"true"`
	}
	Replica ReplicaConfig
}

type ApiConfig struct {
	Port   uint16 `envconfig:"API_PORT" default:"50051" required:"true"`
	Writer struct {
		Backoff   time.Duration `envconfig:"API_WRITER_BACKOFF" default:"10s" required:"true"`
		BatchSize uint32        `envconfig:"API_WRITER_BATCH_SIZE" default:"16" required:"true"`
		Cache     WriterCacheConfig
		Uri       string `envconfig:"API_WRITER_URI" default:"resolver:50051" required:"true"`
	}
	UserAgent string `envconfig:"API_USER_AGENT" default:"Awakari" required:"true"`
	GroupId   string `envconfig:"API_GROUP_ID" default:"default" required:"true"`
	Events    EventsConfig
}

type EventsConfig struct {
	Uri        string `envconfig:"API_EVENTS_URI" default:"events:50051" required:"true"`
	Connection struct {
		Count struct {
			Init uint32 `envconfig:"API_EVENTS_CONN_COUNT_INIT" default:"1" required:"true"`
			Max  uint32 `envconfig:"API_EVENTS_CONN_COUNT_MAX" default:"100" required:"true"`
		}
		IdleTimeout time.Duration `envconfig:"API_EVENTS_CONN_IDLE_TIMEOUT" default:"15m" required:"true"`
	}
	Source string `envconfig:"API_EVENTS_SOURCE" default:"https://awakari.com/pub.html?srcType=sse" required:"true"`
	Limit  uint32 `envconfig:"API_EVENTS_LIMIT" default:"1000" required:"true"`
	Topics TopicsConfig
}

type TopicsConfig struct {
	Mastodon string `envconfig:"API_EVENTS_TOPIC_MASTODON" default:"source-sse-mastodon" required:"true"`
}

type DbConfig struct {
	Uri      string `envconfig:"DB_URI" default:"mongodb://localhost:27017/?retryWrites=true&w=majority" required:"true"`
	Name     string `envconfig:"DB_NAME" default:"sources" required:"true"`
	UserName string `envconfig:"DB_USERNAME" default:""`
	Password string `envconfig:"DB_PASSWORD" default:""`
	Table    struct {
		Name      string        `envconfig:"DB_TABLE_NAME" default:"sse" required:"true"`
		Retention time.Duration `envconfig:"DB_TABLE_RETENTION" default:"2160h" required:"true"`
		Shard     bool          `envconfig:"DB_TABLE_SHARD" default:"true"`
	}
	Tls struct {
		Enabled  bool `envconfig:"DB_TLS_ENABLED" default:"false" required:"true"`
		Insecure bool `envconfig:"DB_TLS_INSECURE" default:"false" required:"true"`
	}
}

type SseConfig struct {
	StreamTimeout time.Duration `envconfig:"SSE_STREAM_TIMEOUT" default:"5m" required:"true"`
	Type          string        `envconfig:"SSE_TYPE" required:"true" default:"com_awakari_sse_v1"`
}

type ReplicaConfig struct {
	Count uint32 `envconfig:"REPLICA_COUNT" required:"true"`
	Name  string `envconfig:"REPLICA_NAME" required:"true"`
}

type WriterCacheConfig struct {
	Size uint32        `envconfig:"API_WRITER_CACHE_SIZE" default:"100" required:"true"`
	Ttl  time.Duration `envconfig:"API_WRITER_CACHE_TTL" default:"24h" required:"true"`
}

func NewConfigFromEnv() (cfg Config, err error) {
	err = envconfig.Process("", &cfg)
	return
}
