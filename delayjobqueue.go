package delayQ

import (
	"errors"
	"github.com/caigoumiao/cronSchedule"
)

type DelayQ struct {
	// 定时任务调度器
	scheduler *cronSchedule.Scheduler

	// 延时任务执行器工厂
	jobExecutorFactory map[string]*jobExecutor

	// redis客户端
	redisCli *redisClient

	// 支持日志自定义
	logger Logger
}

var (
	delayQ *DelayQ

	// errors
	ErrorsDelayQRegisterIDDuplicate = errors.New("your job id has been used")
)

func initDelayQ(conf DelayQConf) {
	delayQ = new(DelayQ)
	// 检查配置
	checkConf(conf)
	// 初始化Logger
	delayQ.logger = defaultLogger
	// 初始化定时任务调度器
	sche := cronSchedule.New()
	sche.Register([]int{}, 1, DelayQCronJob{})
	delayQ.scheduler = sche
	// 初始化任务执行器工厂
	delayQ.jobExecutorFactory = make(map[string]*jobExecutor)
	// 初始化redis
	delayQ.redisCli = getRedisCli(conf.Redis)
	delayQ.logger.InfoF("Initialization of DelayQ completed......")
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
	dq.logger.InfoF("DelayQ start background......")
}

// 注册任务执行器
func (dq *DelayQ) Register(action JobBaseAction) error {
	jobID := action.ID()
	if _, ok := dq.jobExecutorFactory[jobID]; ok {
		dq.logger.ErrorF("DelayQ register job[ID=%s] err=%v", action.ID(), ErrorsDelayQRegisterIDDuplicate)
		return ErrorsDelayQRegisterIDDuplicate
	} else {
		dq.jobExecutorFactory[jobID] = &jobExecutor{
			ID:     jobID,
			action: action,
		}
		dq.logger.InfoF("DelayQ register job[ID=%s]", jobID)
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

// 自定义Logger
func (dq *DelayQ) SetLogger(logger Logger) {
	dq.logger = logger
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
