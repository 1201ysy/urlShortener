package redis

import (
	"fmt"
	"strconv"
	"urlShortener/shortener"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type redisRepository struct{
	client *redis.Client
}

// Start new redis client ; not the RedirectRepository struct
func newRedisClient(redisURL string) (*redis.Client, error){
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Create and return new RedirectRepository for redis
func NewRedisRepository(redisURL string) (shortener.RedirectRepository, error){
	repo := &redisRepository{}
	client, err := newRedisClient(redisURL)
	if err != nil{
		return nil, errors.Wrap(err, "repository.NewRedisRepository")
	}
	repo.client = client
	return repo, nil
}

// Generate keys to be used for redis data
func (r *redisRepository ) generateKey(code string) string{
	return fmt.Sprintf("redirect:%s",code)
}

// Implements RedirectRepository interfaces
// Find redirects from redis keys
func (r *redisRepository ) Find(code string) (*shortener.Redirect, error){
	redirect := &shortener.Redirect{}
	key := r.generateKey(code)
	data, err := r.client.HGetAll(key).Result()
	if err != nil{
		return nil, errors.Wrap(err, "repository.Redirect.Find")
	}
	if len(data) == 0 {
		return nil, errors.Wrap(shortener.ErrRedirectNotFound, "repository.Redirect.Find")
	}
	createdAt, err := strconv.ParseInt(data["created_at"], 10, 64)
	if err != nil{
		return nil, errors.Wrap(err, "repository.Redirect.Find")
	}
	redirect.Code = data["code"]
	redirect.URL = data["url"]
	redirect.CreatedAt = createdAt
	return redirect, nil
}

// Implements RedirectRepository interfaces
// Store redirects into redis key
func (r *redisRepository) Store(redirect *shortener.Redirect) error {
	key := r.generateKey(redirect.Code)
	data := map[string]interface{}{
		"code":       redirect.Code,
		"url":        redirect.URL,
		"created_at": redirect.CreatedAt,
	}
	_, err := r.client.HMSet(key, data).Result()
	if err != nil {
		return errors.Wrap(err, "repository.Redirect.Store")
	}
	return nil
}
