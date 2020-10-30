package delayQ

import (
	"fmt"
	"github.com/go-redis/redis"
)

var redisCli *redis.Client

func initRedis(host string, port int) {
	client := redis.NewClient(
		&redis.Options{
			Addr: fmt.Sprintf("%s:%d", host, port),
		},
	)

	// 2. 检查是否可以访问（ping）
	if err := checkRedis(client); err != nil {
		fmt.Println(err.Error())
		panic("failed to connect to redis")
	}
	redisCli = client
}

func checkRedis(c *redis.Client) error {
	_, err := c.Ping().Result()
	return err
}

func getRedisCli(conf RedisConf) *redis.Client {
	if redisCli == nil {
		initRedis(conf.Host, conf.Port)
	}
	return redisCli
}
