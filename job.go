package delayQ

// 任务的行为
// 每个任务都需要实现这个接口
type JobBaseAction interface {
	// 任务的唯一标识
	ID() string

	// 任务的执行体
	Execute(args []interface{}) error
}

// 任务的执行器
type jobExecutor struct {
	ID     string
	action JobBaseAction
}

// 延迟队列的单个消息体
type DelayJobMsg struct {
	// 任务唯一标识
	JobID string

	// 任务需要延迟的时间
	// 以秒为单位
	DelayTime int

	// 任务执行需要的参数
	Args []interface{}
}
