# delayQ
delayQ 是基于 Golang 实现的延时队列。

目前常见的延时队列实现方式主要有三种：
1. 基于Redis Zset, 以时间戳作为Score, 主动轮询小于当前的时间的元素
2. 基于RabbitMQ 死信队列 + TTL 实现
3. 基于TimeWheel 时间轮算法

delayQ 采用的是第一种实现方式，这种方法理解起来最为简单，最重要的是能够快速落地。

## 安装
````
go get -u -v github.com/caigoumiao/delayQ
````
推荐使用go.mod
<br>
````
require github.com/caigoumiao/delayQ latest
````
## 简单使用
1、得到 DelayQ 主实例

实例化 DelayQ 需要提供 DelayQConf 的结构体参数，这是DelayQ需要用到的一份配置参数，部分为必填项，具体信息可见DelayQConf的注释说明。

```go
conf := DelayQConf{
        Redis: RedisConf{
            Host: "127.0.0.1",
            Port: 6379,
        },
    }
dq := New(conf)
```

DelayQ支持自定义日志实现，如果不提供则选用默认实现。
```go
type Logger interface {
	InfoF(format string, args ...interface{})
	ErrorF(format string, args ...interface{})
}

// logger 为用户自己的 Logger 实现
dq.SetLogger(logger)
```

2、实现JobBaseAction 接口，定义自己的任务行为

> 任务行为：指的是某一种任务的具体执行行为。<br>
> 任务参数：指的是执行任务行为需要的参数。

举个例子：
+ 给指定手机号列表的用户发送短信
    + 任务参数是：指定的手机号列表
    + 任务行为是：给指定用户发送短信
+ 将指定用户的会员降权
    + 任务参数是：指定的用户列表
    + 任务行为是：把指定用户的会员降权

JobBaseAction 接口有两个需要实现的接口:
+ `ID()` 任务行为的唯一标识
+ `Execute(args []interface{})` 任务的具体行为

````go
// 发送短信的任务行为
type JobActionSMS struct{}

func (JobActionSMS) ID() string {
    return "JobActionSMS"
}

func (JobActionSMS) Execute(args []interface{}) error {
    for _,arg := range args {
        if phoneNumber,ok := arg.(string);ok {
            fmt.Printf("sending sms to %s", phoneNumber)
        } else {
            // 
        }
    }
    return nil
}

// 用户会员降权的任务行为
type JobActionUserDownRight struct{}

func (JobActionUserDownRight) ID() string {
    return "JobActionUserDownRight"
}

func (JobActionUserDownRight) Execute(args []interface{}) error {
    uids := make([]int64, 0)
    for i:=0;i<len(args);i++ {
        if id,ok := args[i].(int64);ok {
            uids = append(uids, id)
        }
    }
    return UserBO.DownRights(uids)
}
````

3、注册任务行为并开启延时任务队列

将自己需要用到的任务行为注册到 DelayQ。
```go
dq.Register(JobActionSMS{})
dq.Register(JobActionUserDownRight{})
dq.StartBackground()
```

4、添加需要延时执行任务

```go
// 2分钟后给手机号为1302xxx9421的用户发条短信
dq.AddDelay(DelayJobMsg{
    JobID:     "JobActionSMS",
    DelayTime: 2*60,
    Args:      []interface{}{"1302xxxx9421"},
})

// 5小时后将用户{7327,8293729,4397892}的会员降权
dq.AddDelay(DelayJobMsg{
    JobID:     "JobActionUserDownRight",
    DelayTime: 5*3600,
    Args:      []interface{}{7327,8293729,4397892},
})
```

DelayJobMsg 结构说明：
```go
// 延迟队列的单个任务体
type DelayJobMsg struct {
	// 任务行为唯一标识
	JobID string

	// 此任务需要延迟的时间
	// 以秒为单位
	DelayTime int

	// 任务执行需要的参数
	Args []interface{}
}
```

## 致谢
相遇是缘！感恩🙏🙏🙏

如果你喜欢本项目或本项目有帮助到你，希望你可以帮忙 star 一下。

如果你有任何意见或建议，欢迎提 issue 或联系我本人。联系方式如下：
+ 微信：wo4qiaoba
