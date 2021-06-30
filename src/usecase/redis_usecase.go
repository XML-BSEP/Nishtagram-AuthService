package usecase

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	logger "github.com/jelena-vlajkov/logger/logger"
	"time"
)

type RedisUsecase interface {
	AddKeyValueSet(context context.Context, key string, value interface{},  expiration time.Duration) error
	GetValueByKey(context context.Context, key string) ([]byte, error)
	DeleteValueByKey(context context.Context, key string) error
	ExistsByKey(context context.Context, key string) bool
	ScanKeyByPattern(ctx context.Context, pattern string) ([]string, error)
}

type redisUsecase struct {
	RedisClient *redis.Client
	logger *logger.Logger
}


func NewRedisUsecase(r *redis.Client, logger *logger.Logger) RedisUsecase{
	return &redisUsecase{RedisClient: r, logger: logger}
}

func (r *redisUsecase) GetValueByKey(context context.Context, key string) ([]byte, error) {
	return r.RedisClient.Get(context, key).Bytes()

}

func (r *redisUsecase) AddKeyValueSet(context context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.RedisClient.Ping(context).Err()
	if err != nil {
		r.logger.Logger.Errorf("error while adding to redis, error: %v\n", err)
	}
	return r.RedisClient.Set(context, key, value, expiration).Err()
}

func (r *redisUsecase) DeleteValueByKey(context context.Context, key string) error {
	err := r.RedisClient.Del(context, key).Err()
	if err != nil {
		r.logger.Logger.Errorf("error while deleting value from redis, error: %b\n", err)
	}
	return err
}

func (r *redisUsecase) ExistsByKey(context context.Context, key string) bool {
	res := r.RedisClient.Exists(context, key).Val()
	if res == 0 {
		return false
	}
	return true
}

func (r *redisUsecase) ScanKeyByPattern(ctx context.Context, pattern string) ([]string, error) {

	r.logger.Logger.Infof("scanning redis for stories to show by pattern %v\n", pattern)
	var _, cursor uint64
	var n int
	var keysLength []string
	keysLength, _, _ = r.RedisClient.Scan(ctx, cursor, "*", 0).Result()

	numFor := len(keysLength)
	for {

		var keys []string
		var err error
		keys, cursor, err = r.RedisClient.Scan(context.Background(), cursor, pattern, int64(numFor)).Result()
		if err != nil {
			panic(err)
		}
		n += len(keys)
		if len(keys) > 0 {
			return keys, nil
		}

		if cursor == 0 {
			break
		}
	}
	r.logger.Logger.Warnf("no stories to show by pattern %v\n", pattern)
	return nil, fmt.Errorf("server error")

}