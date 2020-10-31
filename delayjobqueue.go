package delayQ

import (
	"errors"
	"github.com/caigoumiao/cronSchedule"
)

type DelayQ struct {
	scheduler          *cronSchedule.Scheduler
	jobExecutorFactory map[string]*jobExecutor
	redisCli           *redisClient
}

var (
	delayQ *DelayQ

	// errors
	ErrorsDelayQRegisterIDDuplicate = errors.New("your job id has been used")
)

func initDelayQ(conf DelayQConf) *DelayQ {
	dq := new(DelayQ)
	// 检查配置
	checkConf(conf)
	// 初始化定时任务调度器
	sche := cronSchedule.New()
	sche.Register([]int{}, 1, DelayQCronJob{})
	dq.scheduler = sche
	// 初始化任务执行器工厂
	dq.jobExecutorFactory = make(map[string]*jobExecutor)
	// 初始化redis
	dq.redisCli = getRedisCli(conf.Redis)
	return dq
}

// 检查配置
// 目前就是检查必填项，非必填项补充为默认值
// todo: 写法不好👎
func checkConf(conf DelayQConf) {
	if conf.Redis.KeyPrefix == "" {
		conf.Redis.KeyPrefix = defaultDelayQKeyPrefix
	}
	if conf.Redis.ZSetBatchLimit == 0 {
		conf.Redis.ZSetBatchLimit = 1000
	}
}

func New(conf DelayQConf) *DelayQ {
	if delayQ == nil {
		initDelayQ(conf)
	}
	return delayQ
}

// 启动延迟队列后台
func (dq *DelayQ) StartBackground() {
	dq.scheduler.Start()
}

// 注册任务执行器
func (dq *DelayQ) Register(action JobBaseAction) error {
	jobID := action.ID()
	if _, ok := dq.jobExecutorFactory[jobID]; ok {
		return ErrorsDelayQRegisterIDDuplicate
	} else {
		dq.jobExecutorFactory[jobID] = &jobExecutor{
			ID:     jobID,
			action: action,
		}
	}
	return nil
}

// 往队列添加延迟任务
func (dq *DelayQ) AddDelay(job DelayJobMsg) error {
	return dq.redisCli.ZAdd(job)
}

// 获取所有可用的jobId
func (dq *DelayQ) availableJobIDs() []string {
	var IDs []string
	for k, _ := range dq.jobExecutorFactory {
		IDs = append(IDs, k)
	}
	return IDs
}

// 延迟队列的配置项
type (
	// 总配置
	DelayQConf struct {
		Redis RedisConf
	}
	// redis 配置项
	RedisConf struct {
		// 主机地址（*）
		Host string

		// 端口号（*）
		Port int

		// 如果不设置，则默认为delayQ
		KeyPrefix string

		// ZSet批量限制的条数，默认为1000
		ZSetBatchLimit int64
	}
)

// 延迟后台的定时任务，每秒运行一次
// 拉取redis中的到期任务执行
type DelayQCronJob struct{}

func (DelayQCronJob) Name() string {
	return "DelayQCron"
}

func (DelayQCronJob) Process() error {
	IDs := delayQ.availableJobIDs()
	return delayQ.redisCli.BatchHandle(IDs)
}

func (DelayQCronJob) IfActive() bool {
	return true
}

func (DelayQCronJob) IfReboot() bool {
	return true
}
