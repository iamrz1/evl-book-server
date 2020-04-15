package db

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"log"
)

var redisClient RedisClient

type RedisClient struct {
	*redis.Client
}

const (
	RedisNilErr = "redis: nil"
)

// Setup setups the redis client instance with the requied infos
func (r *RedisClient) SetupMyRedis() {
	r.Client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.db_url") + ":" + viper.GetString("redis.db_port"),
		Password: viper.GetString("redis.db_password"),
		DB:       viper.GetInt("redis.db"),
	})
}

// ConnectRedis starts the redis connection as a client
func InitRedis() {
	log.Println("setup redis client:")
	redisClient.SetupMyRedis()
}

func IsRedisUp() bool {
	pong, err := redisClient.Ping().Result()
	if err != nil {
		log.Println("failed to setup db")
		return false
	}
	if pong != "PONG" {
		log.Println("ping failed")
		return false
	}
	return true
}

// SetJsonValues set json values against uid in redis instance
func SetJsonValues(key string, json []byte) error {
	return redisClient.Set(key, json, 0).Err()
}
func SetSingleValue(key string, value string) error {
	return redisClient.Set(key, value, 0).Err()
}

// GetByteValues set json values against uid in redis instance
func GetByteValues(key string) ([]byte, error) {
	val, err := redisClient.Get(key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

func GetSingleValue(key string) (string, error) {
	val, err := redisClient.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func ReplaceKey(key string, newKey string) error {
	return redisClient.Rename(key, newKey).Err()
}
func GetClient() *RedisClient {
	return &redisClient
}

func CloseRedis() {
	log.Println("closed redis client")
	err := redisClient.Close()
	if err != nil {
		log.Println("couldn't close redis server")
	}
}
