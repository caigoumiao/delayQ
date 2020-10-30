package delayQ

import (
	"errors"
	"github.com/caigoumiao/cronSchedule"
	"github.com/go-redis/redis"
)

type DelayQ struct {
	scheduler          *cronSchedule.Scheduler
	jobExecutorFactory map[string]*jobExecutor
	redisCli           *redis.Client
}

var (
	delayQ *DelayQ

	// errors
	ErrorsDelayQRegisterIDDuplicate = errors.New("your job id has been used")
)

func initDelayQ(conf DelayQConf) *DelayQ {
	dq := new(DelayQ)
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
func (*DelayQ) AddDelay(job DelayJobMsg) error {
	return nil
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
	}
)

// 延迟后台的定时任务，每秒运行一次
// 拉取redis中的到期任务执行
type DelayQCronJob struct{}

func (DelayQCronJob) Name() string {
	return "DelayQCron"
}

func (DelayQCronJob) Process() error {

	return nil
}

func (DelayQCronJob) IfActive() bool {
	return true
}

func (DelayQCronJob) IfReboot() bool {
	return true
}
