package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alaajili/task-scheduler/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	cleanup := func() {
		mockDB.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	}

	return mockDB, mock, cleanup
}

func mockSqlOpen(t *testing.T, mockDB *sql.DB, returnErr error) func() {
	t.Helper()
	originalSqlOpen := sqlOpen

	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		if returnErr != nil {
			return nil, returnErr
		}
		return mockDB, nil
	}

	return func() {
		sqlOpen = originalSqlOpen
	}
}

func getTestDBConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     "localhost",
		Port:	  5432,
		User:	  "testuser",
		Password: "testpassword",
		DBName:   "testdb",
		SSLMode:  "disable",
		MaxConns: 10,
		MaxIdle:  5,
	}
}

func TestNewPostgresDB_Success(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	
	unmock := mockSqlOpen(t, mockDB, nil)
	defer unmock()

	mock.ExpectPing()

	db, err := NewPostgresDB(getTestDBConfig())

	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestNewPostgresDB_OpenError(t *testing.T) {
	expectedErr := errors.New("failed to connect")
	unmock := mockSqlOpen(t, nil, expectedErr)
	defer unmock()

	db, err := NewPostgresDB(getTestDBConfig())

	require.Error(t, err)
	require.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to open database")
}

func TestNewPostgresDB_PingError(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	
	unmock := mockSqlOpen(t, mockDB, nil)
	defer unmock()
	mock.ExpectPing().WillReturnError(errors.New("ping failed"))

	db, err := NewPostgresDB(getTestDBConfig())

	require.Error(t, err)
	require.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestNewPostgresDB_ConnectionPoolSettings(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	
	unmock := mockSqlOpen(t, mockDB, nil)
	defer unmock()

	mock.ExpectPing()

	cfg := getTestDBConfig()
	cfg.MaxConns = 20
	cfg.MaxIdle = 10

	db, err := NewPostgresDB(cfg)

	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, db.DB)
	require.Equal(t, 20, db.DB.Stats().MaxOpenConnections)
}

func TestNewPostgresDB_DSNFormat(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	var capturedDSN string
	originalSQLOpen := sqlOpen
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		capturedDSN = dataSourceName
		return mockDB, nil
	}
	defer func() { sqlOpen = originalSQLOpen }()

	mock.ExpectPing()

	cfg := config.DatabaseConfig{
		Host:     "testhost",
		Port:     5433,
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		SSLMode:  "require",
		MaxConns: 10,
		MaxIdle:  5,
	}

	db, err := NewPostgresDB(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, db)
	
	expectedDSN := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=require"
	assert.Equal(t, expectedDSN, capturedDSN)
}

func TestDB_Close(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	db := &DB{DB: mockDB}

	mock.ExpectClose()
	
	err = db.Close()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_HealthCheck_Success(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()
	
	db := &DB{DB: mockDB}
	mock.ExpectPing()

	err := db.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestDB_HealthCheck_Error(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	db := &DB{DB: mockDB}
	mock.ExpectPing().WillReturnError(errors.New("ping error"))
	
	err := db.HealthCheck(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping error")
}

func TestDB_HealthCheck_WithTimeout(t *testing.T) {
	mockDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	db := &DB{DB: mockDB}
	mock.ExpectPing()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := db.HealthCheck(ctx)
	assert.NoError(t, err)
}
