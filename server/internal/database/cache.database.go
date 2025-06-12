package database

import (
	"context"
	"fmt"
	"server/config"
	"server/internal/logger"

	"github.com/valkey-io/valkey-go"
)

const (
	GENERAL_CACHE_INDEX = iota
	SESSION_CACHE_INDEX
	USER_CACHE_INDEX
	EVENTS_CACHE_INDEX
)

func (s *DB) initializeCacheDB(config config.Config) error {
	log := s.log.Function("initializeCacheDB")
	log.Info("initializing cache database")

	address := config.DatabaseCacheAddress
	port := config.DatabaseCachePort
	if address == "" || port == 0 {
		return log.Errorf("failed to initialize cache database", "address or port is empty")
	}

	var cacheDB Cache

	var err error
	cacheDB.General, err = valkey.NewClient(
		valkey.ClientOption{
			InitAddress: []string{fmt.Sprintf("%s:%d", address, port)},
			SelectDB:    GENERAL_CACHE_INDEX,
		},
	)
	if err != nil || testCacheDB(cacheDB.General, log) != nil {
		return log.Err("failed to create and test general valkey client", err)
	}

	cacheDB.Session, err = valkey.NewClient(
		valkey.ClientOption{
			InitAddress: []string{fmt.Sprintf("%s:%d", address, port)},
			SelectDB:    SESSION_CACHE_INDEX,
		},
	)
	if err != nil || testCacheDB(cacheDB.Session, log) != nil {
		return log.Err("failed to create and test session valkey client", err)
	}

	cacheDB.User, err = valkey.NewClient(
		valkey.ClientOption{
			InitAddress: []string{fmt.Sprintf("%s:%d", address, port)},
			SelectDB:    USER_CACHE_INDEX,
		},
	)
	if err != nil || testCacheDB(cacheDB.User, log) != nil {
		return log.Err("failed to create and test user valkey client", err)
	}

	cacheDB.Events, err = valkey.NewClient(
		valkey.ClientOption{
			InitAddress: []string{fmt.Sprintf("%s:%d", address, port)},
			SelectDB:    EVENTS_CACHE_INDEX,
		},
	)
	if err != nil || testCacheDB(cacheDB.Events, log) != nil {
		return log.Err("failed to create and test events valkey client", err)
	}

	s.Cache = cacheDB

	return nil
}

func testCacheDB(client valkey.Client, log logger.Logger) error {
	log = log.Function("testCacheDB")
	ctx := context.Background()
	err := client.Do(ctx, client.B().Ping().Build()).Error()
	if err != nil {
		return log.Err("failed to ping valkey", err)
	}

	return nil
}

// func valueToString[T any](value T) (string, error) {
// 	if str, ok := any(value).(string); ok {
// 		return str, nil
// 	}
//
// 	bytes, err := json.Marshal(value)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(bytes), nil
// }
