package usecase

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisUsecase interface {
	AddKeyValueSet(context context.Context, key string, value interface{},  expiration time.Duration) error
	GetValueByKey(context context.Context, key string) ([]byte, error)
	DeleteValueByKey(context context.Context, key string) error
	ExistsByKey(context context.Context, key string) bool
}

type redisUsecase struct {
	RedisClient *redis.Client
}

func NewRedisUsecase(r *redis.Client) RedisUsecase{
	return &redisUsecase{RedisClient: r}
}

func (r *redisUsecase) GetValueByKey(context context.Context, key string) ([]byte, error) {
	return r.RedisClient.Get(context, key).Bytes()

}

func (r *redisUsecase) AddKeyValueSet(context context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.RedisClient.Ping(context).Err()
	if err != nil {
		fmt.Println(err)
	}
	return r.RedisClient.Set(context, key, value, expiration).Err()
}

func (r *redisUsecase) DeleteValueByKey(context context.Context, key string) error {
	return r.RedisClient.Del(context, key).Err()
}

func (r *redisUsecase) ExistsByKey(context context.Context, key string) bool {
	res := r.RedisClient.Exists(context, key).Val()
	if res == 0 {
		return false
	}
	return true
}
