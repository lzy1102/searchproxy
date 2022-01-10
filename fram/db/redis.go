package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"reflect"
	"searchproxy/fram/config"
	"searchproxy/fram/utils"
	"sync"
	"time"
)

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type RedisClient struct {
	cfg *RedisConfig
	cli *redis.Client
	ctx context.Context
}

var ronce sync.Once
var rds *RedisClient

func RedisInstance() *RedisClient {
	ronce.Do(func() {
		rds = new(RedisClient)
		var cfg RedisConfig
		config.Install().Get("cache", &cfg)
		rds.cfg = &cfg
		rdb := redis.NewClient(&redis.Options{
			Addr:     rds.cfg.Addr,
			Password: rds.cfg.Password,
			DB:       rds.cfg.DB,
		})
		ctx := context.Background()
		_, err := rdb.Ping(ctx).Result()
		utils.FatalAssert(err)
		rds.cli = rdb
		rds.ctx = ctx
	})
	return rds
}

func NewRedis(cfg *RedisConfig) *RedisClient {
	rc := new(RedisClient)
	rc.cfg = cfg
	rdb := redis.NewClient(&redis.Options{
		Addr:     rc.cfg.Addr,
		Password: rc.cfg.Password,
		DB:       rc.cfg.DB,
	})
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	utils.FatalAssert(err)
	rc.cli = rdb
	rc.ctx = ctx
	return rc
}
func (c RedisClient) Close() {
	c.cli.Close()
}

func (c RedisClient) Del(key string) {
	err := c.cli.Del(c.ctx, key).Err()
	utils.FatalAssert(err)
}

func (c RedisClient) Set(key string, value interface{}, expiration time.Duration) {
	if reflect.TypeOf(value).String() == reflect.TypeOf(map[string]interface{}{}).String() ||
		reflect.TypeOf(value).String() == reflect.TypeOf([]interface{}{}).String() ||
		reflect.TypeOf(value).String() == reflect.TypeOf(struct{}{}).String() {
		value, _ = json.Marshal(value)
	}
	err := c.cli.Set(c.ctx, key, value, expiration).Err()
	utils.FatalAssert(err)
}

func (c RedisClient) Get(key string) (string, error) {
	exint := c.Exists(key)
	if exint == int64(0) {
		return "", fmt.Errorf("exits err")
	}
	res, err := c.cli.Get(c.ctx, key).Result()
	return res, err
}
func (c RedisClient) GetInt64(key string) int64 {
	exint := c.Exists(key)
	if exint == int64(0) {
		return int64(0)
	}
	res, err := c.cli.Get(c.ctx, key).Int64()
	utils.FatalAssert(err)
	return res
}

func (c RedisClient) GetInt(key string) int {
	exint := c.Exists(key)
	if exint == int64(0) {
		return 0
	}
	res, err := c.cli.Get(c.ctx, key).Int()
	utils.FatalAssert(err)
	return res
}

func (c RedisClient) Exists(key ...string) int64 {
	ex := c.cli.Exists(c.ctx, key...)
	exint, err := ex.Result()
	utils.FatalAssert(err)
	return exint
}

func (c RedisClient) GetFloat32(key string) float32 {
	exint := c.Exists(key)
	if exint == int64(0) {
		return 0
	}
	res, err := c.cli.Get(c.ctx, key).Float32()
	utils.FatalAssert(err)
	return res
}
func (c RedisClient) GetFloat64(key string) float64 {
	exint := c.Exists(key)
	if exint == int64(0) {
		return float64(0)
	}
	res, err := c.cli.Get(c.ctx, key).Float64()
	utils.FatalAssert(err)
	return res
}
func (c RedisClient) GetUint64(key string) uint64 {
	exint := c.Exists(key)
	if exint == int64(0) {
		return uint64(0)
	}
	res, err := c.cli.Get(c.ctx, key).Uint64()
	utils.FatalAssert(err)
	return res
}

func (c RedisClient) Append(key string, value string) {
	c.cli.Append(c.ctx, key, value)
}

func (c RedisClient) MGet(key ...string) interface{} {
	exint := c.Exists(key...)
	if exint == int64(0) {
		return ""
	}
	res, err := c.cli.MGet(c.ctx, key...).Result()
	utils.FatalAssert(err)
	return res
}

func (c RedisClient) Decrby(key string, decrement int64) {
	c.cli.DecrBy(c.ctx, key, decrement)
}

func (c RedisClient) Decr(key string) {
	c.cli.Decr(c.ctx, key)
}

func (c RedisClient) IncrByFloat(key string, decrement float64) {
	c.cli.IncrByFloat(c.ctx, key, decrement)
}

func (c RedisClient) IncrBy(key string, value int64) {
	c.cli.IncrBy(c.ctx, key, value)
}

func (c RedisClient) Keys(key string) interface{} {
	res, err := c.cli.Keys(c.ctx, key).Result()
	utils.FatalAssert(err)
	return res
}

func (c RedisClient) MSet(values ...interface{}) {
	c.cli.MSet(c.ctx, values)
}

func (c RedisClient) Scan(match string) []string {
	var cursor uint64
	var n int
	var results []string
	for {
		var keys []string
		var err error
		//*扫描所有key，每次20条
		keys, cursor, err = c.cli.Scan(c.ctx, cursor, match, 20).Result()
		if err != nil {
			panic(err)
		}
		n += len(keys)
		//var value string
		//for _, key := range keys {
		//	value, err = c.cli.Get(c.ctx, key).Result()
		//}
		results = append(results, keys...)
		if cursor == 0 {
			break
		}
	}
	return results
}
