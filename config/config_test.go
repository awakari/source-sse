package config

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	os.Setenv("API_WRITER_BACKOFF", "23h")
	os.Setenv("API_WRITER_URI", "writer:56789")
	os.Setenv("LOG_LEVEL", "4")
	os.Setenv("REPLICA_COUNT", "2")
	os.Setenv("REPLICA_NAME", "source-sse-1")
	os.Setenv("API_TOKEN_INTERNAL", "foo")
	cfg, err := NewConfigFromEnv()
	assert.Nil(t, err)
	assert.Equal(t, 23*time.Hour, cfg.Api.Writer.Backoff)
	assert.Equal(t, "writer:56789", cfg.Api.Writer.Uri)
	assert.Equal(t, slog.LevelWarn, slog.Level(cfg.Log.Level))
}
