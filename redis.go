package delayQ

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"sync"
	"time"
)

type redisClient struct {
	keyPrefix  string
	batchLimit int64
	conn       *redis.Client
}

var (
	redisCli               *redisClient
	defaultDelayQKeyPrefix = "delayQ"
)

func (cli *redisClient) ZAdd(msg DelayJobMsg) error {
	key := cli.formatKey(msg.JobID)
	member, _ := json.Marshal(msg)
	return cli.conn.ZAdd(key, redis.Z{
		Score:  float64(time.Now().Unix() + int64(msg.DelayTime)),
		Member: string(member),
	}).Err()
}

func (cli *redisClient) BatchHandle(IDs []string) error {
	var wg = sync.WaitGroup{}
	wg.Add(len(IDs))
	for _, name := range IDs {
		key := cli.formatKey(name)
		go func(key string) {
			batch, lastScore, err := cli.getBatch(key)
			if err != nil {
				// 日志报警
			} else {
				for _, item := range batch {
					var m DelayJobMsg
					if err := json.Unmarshal([]byte(item.Member.(string)), &m); err != nil {
						// log
						continue
					}
					// 寻找任务对应的执行器
					if executor, ok := delayQ.jobExecutorFactory[m.JobID]; !ok {
						// log
						continue
					} else {
						// 延迟任务开始执行
						executor.action.Execute(m.Args)
					}
				}
			}
			defer func() {
				if err != nil || len(batch) != 0 {
					cli.clearBatch(key, lastScore)
				}
				wg.Done()
			}()
		}(key)
	}
	wg.Wait()
	return nil
}

func (cli *redisClient) getBatch(key string) (batch []redis.Z, lastScore float64, err error) {
	batch, err = cli.conn.ZRangeByScoreWithScores(key, redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.Itoa(int(time.Now().Unix())),
		Offset: 0,
		Count:  cli.batchLimit,
	}).Result()
	if err != nil || len(batch) == 0 {
		return
	}
	lastScore = batch[len(batch)-1].Score
	batch, err = cli.conn.ZRangeByScoreWithScores(key, redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.Itoa(int(lastScore)),
		Offset: 0,
		Count:  cli.batchLimit,
	}).Result()
	return
}

func (cli *redisClient) clearBatch(key string, lastScore float64) error {
	return cli.conn.ZRemRangeByScore(key, "-inf", strconv.Itoa(int(lastScore))).Err()
}

func (cli *redisClient) formatKey(name string) string {
	return fmt.Sprintf("%s:%s", cli.keyPrefix, name)
}

func initRedis(conf RedisConf) {
	client := redis.NewClient(
		&redis.Options{
			Addr: fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		},
	)

	// 2. 检查是否可以访问（ping）
	if err := checkRedis(client); err != nil {
		delayQ.logger.ErrorF("DelayQ failed to init redis err=%s", err.Error())
		panic("failed to connect to redis")
	}
	redisCli = &redisClient{
		keyPrefix:  conf.KeyPrefix,
		batchLimit: conf.ZSetBatchLimit,
		conn:       client,
	}
	delayQ.logger.InfoF("DelayQ init redis OK!")
}

func checkRedis(c *redis.Client) error {
	_, err := c.Ping().Result()
	return err
}

func getRedisCli(conf RedisConf) *redisClient {
	if redisCli == nil {
		initRedis(conf)
	}
	return redisCli
}
