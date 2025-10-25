package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig("../../config.example.yaml")
	assert.NoError(t, err)

	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 10*time.Second, config.Server.ShutdownTimeout)

	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "postgres", config.Database.User)
	assert.Equal(t, "postgres", config.Database.Password)
	assert.Equal(t, "taskscheduler", config.Database.DBName)
	assert.Equal(t, "disable", config.Database.SSLMode)
	assert.Equal(t, 25, config.Database.MaxConns)
	assert.Equal(t, 5, config.Database.MaxIdle)

	assert.Equal(t, "localhost", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, "", config.Redis.Password)
	assert.Equal(t, 0, config.Redis.DB)

	assert.Equal(t, []string{"localhost:2379"}, config.Etcd.Endpoints)
	assert.Equal(t, 5*time.Second, config.Etcd.Timeout)

	assert.Equal(t, 10*time.Second, config.Worker.HeartbeatInterval)
	assert.Equal(t, 1*time.Second, config.Worker.TaskPollInterval)
	assert.Equal(t, 5*time.Minute, config.Worker.TaskTimeout)
	assert.Equal(t, 10, config.Worker.MaxConcurrent)
}

func TestDSN(t *testing.T) {
	dbConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "taskscheduler",
		SSLMode:  "disable",
	}

	expectedDSN := "host=localhost port=5432 user=postgres password=postgres dbname=taskscheduler sslmode=disable"
	assert.Equal(t, expectedDSN, dbConfig.DSN())
}
