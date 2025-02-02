package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"muzz-homework/internal/explore/domain"
	"time"
)

type RedisConfig struct {
	Prefix string
	TTL    time.Duration
}

type likersResult struct {
	Likers    []domain.LikerInfo `json:"likers"`
	Timestamp *uint64            `json:"timestamp"`
}

type RedisCache struct {
	redis  *redis.Client
	config RedisConfig
}

func NewRedisCache(redis *redis.Client, config RedisConfig) *RedisCache {
	return &RedisCache{
		redis:  redis,
		config: config,
	}
}

func (r *RedisCache) GetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool) ([]domain.LikerInfo, *uint64, error) {
	timestampValue := uint64(0)
	if timestamp != nil {
		timestampValue = *timestamp
	}

	key := fmt.Sprintf("%s:likers:%s:%d:%t", r.config.Prefix, recipientID, timestampValue, excludeMutual)

	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, nil, err
	}

	var result likersResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, nil, err
	}

	return result.Likers, result.Timestamp, nil
}

func (r *RedisCache) SetLikers(ctx context.Context, recipientID string, timestamp *uint64, excludeMutual bool, likers []domain.LikerInfo, nextTS *uint64) error {
	timestampValue := uint64(0)
	if timestamp != nil {
		timestampValue = *timestamp
	}

	key := fmt.Sprintf("%s:likers:%s:%d:%t", r.config.Prefix, recipientID, timestampValue, excludeMutual)

	result := likersResult{
		Likers:    likers,
		Timestamp: nextTS,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return r.redis.Set(ctx, key, data, r.config.TTL).Err()
}

func (r *RedisCache) GetLikersCount(ctx context.Context, recipientID string) (uint64, error) {
	key := fmt.Sprintf("%s:count:%s", r.config.Prefix, recipientID)
	return r.redis.Get(ctx, key).Uint64()
}

func (r *RedisCache) SetLikersCount(ctx context.Context, recipientID string, count uint64) error {
	key := fmt.Sprintf("%s:count:%s", r.config.Prefix, recipientID)
	return r.redis.Set(ctx, key, count, r.config.TTL).Err()
}
