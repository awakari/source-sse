package config

import (
	"github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	Api   ApiConfig
	Db    DbConfig
	Event EventConfig
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

type EventConfig struct {
	StreamTimeout time.Duration `envconfig:"EVENT_STREAM_TIMEOUT" default:"5m" required:"true"`
	Type          string        `envconfig:"EVENT_TYPE" required:"true" default:"com_awakari_sse_v1"`
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
