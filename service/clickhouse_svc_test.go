package service

import (
	"testing"

	dbClickhouseConf "github.com/hecc-blot/db-clickhouse/config"

	"github.com/stretchr/testify/assert"
)

func TestBuildDsn(t *testing.T) {
	cfg := &dbClickhouseConf.ClickhouseConfig{
		Ip:             "127.0.0.1",
		Port:           9000,
		Username:       "default",
		Password:       "secret",
		Database:       "logs",
		ConnectTimeout: 3,
	}
	assert.Equal(t, "clickhouse://default:secret@127.0.0.1:9000/logs?dial_timeout=3s", buildDsn(cfg))
}
