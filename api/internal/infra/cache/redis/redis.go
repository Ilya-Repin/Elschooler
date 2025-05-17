package redis

import (
	"Elschool-API/internal/config"
	"Elschool-API/internal/domain/models"
	"Elschool-API/internal/infra/cache"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCache struct {
	conn *redis.Client
}

func New(conn *redis.Client) *RedisCache {
	return &RedisCache{conn}
}

func (r *RedisCache) FindToken(ctx context.Context, studID string) (string, error) {
	const op = "infra.cache.FindToken"

	result, err := r.conn.Get(ctx, studID).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("%s: %w", op, cache.ErrTokenNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (r *RedisCache) AddToken(ctx context.Context, studID, jwt string) error {
	const op = "infra.cache.AddToken"

	err := r.conn.Set(ctx, studID, jwt, 72*time.Hour).Err()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisCache) DeleteToken(ctx context.Context, studID string) error {
	const op = "infra.cache.DeleteToken"

	if err := r.conn.Del(ctx, studID).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("%s: %w", op, cache.ErrTokenNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisCache) SaveDayMarks(ctx context.Context, studID string, marks models.DayMarks) error {
	const op = "infra.cache.SaveDayMarks"

	key := studID + ":day_marks:" + marks.Date
	data, err := json.Marshal(marks)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.conn.Set(ctx, key, data, 7*time.Second).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisCache) SaveAverageMarks(ctx context.Context, studID string, marks models.AverageMarks) error {
	const op = "infra.cache.SaveAverageMarks"

	key := studID + ":average_marks:" + string(marks.Period)
	data, err := json.Marshal(marks)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.conn.Set(ctx, key, data, 7*time.Second).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisCache) SaveFinalMarks(ctx context.Context, studID string, marks models.FinalMarks) error {
	const op = "infra.cache.SaveFinalMarks"

	key := studID + ":final_marks"
	data, err := json.Marshal(marks)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	err = r.conn.Set(ctx, key, data, 7*time.Second).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *RedisCache) GetDayMarks(ctx context.Context, studID, date string) (models.DayMarks, error) {
	const op = "infra.cache.GetDayMarks"

	key := studID + ":day_marks:" + date
	result, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.DayMarks{}, fmt.Errorf("%s: %w", op, cache.ErrMarksNotFound)
		}
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	var marks models.DayMarks
	err = json.Unmarshal([]byte(result), &marks)
	if err != nil {
		return models.DayMarks{}, err
	}
	return marks, nil
}

func (r *RedisCache) GetAverageMarks(ctx context.Context, studID string, period int32) (models.AverageMarks, error) {
	const op = "infra.cache.GetAverageMarks"

	key := studID + ":average_marks:" + string(period)
	result, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.AverageMarks{}, fmt.Errorf("%s: %w", op, cache.ErrMarksNotFound)
		}
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	var marks models.AverageMarks
	err = json.Unmarshal([]byte(result), &marks)
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	return marks, nil
}

func (r *RedisCache) GetFinalMarks(ctx context.Context, studID string) (models.FinalMarks, error) {
	const op = "infra.cache.GetFinalMarks"

	key := studID + ":final_marks"
	result, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return models.FinalMarks{}, fmt.Errorf("%s: %w", op, cache.ErrMarksNotFound)
		}
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	var marks models.FinalMarks
	err = json.Unmarshal([]byte(result), &marks)
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	return marks, nil
}

func InitCache(cfg *config.CacheConfig) (*redis.Client, error) {
	rclient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DB:   cfg.Base,
	})

	ctx := context.Background()
	_, err := rclient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rclient, nil
}
