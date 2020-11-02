package delayQ

import (
	"fmt"
	"testing"
	"time"
)

var (
	JobID1 = "jobid1"
	JobID2 = "jobid2"
)

type JobActionSMS struct{}

func (JobActionSMS) ID() string {
	return "JobActionSMS"
}

func (JobActionSMS) Execute(args []interface{}) error {
	for _, arg := range args {
		if phoneNumber, ok := arg.(string); ok {
			fmt.Printf("sending sms to %s", phoneNumber)
		} else {
			//
		}
	}
	return nil
}

type JobActionUserDownRight struct{}

func (JobActionUserDownRight) ID() string {
	return "JobActionUserDownRight"
}

func (JobActionUserDownRight) Execute(args []interface{}) error {
	uids := make([]int64, 0)
	for i := 0; i < len(args); i++ {
		if id, ok := args[i].(int64); ok {
			uids = append(uids, id)
		}
	}
	return nil
}

func TestNew(t *testing.T) {
	conf := DelayQConf{
		Redis: RedisConf{
			Host: "127.0.0.1",
			Port: 6379,
		},
	}
	dq := New(conf)
	dq.Register(JobActionSMS{})
	dq.Register(JobActionUserDownRight{})
	dq.StartBackground()
	time.Sleep(2 * time.Second)
	fmt.Println("add delay")
	dq.AddDelay(DelayJobMsg{
		JobID:     "JobActionSMS",
		DelayTime: 2 * 60,
		Args:      []interface{}{"1302xxxx9421"},
	})
	dq.AddDelay(DelayJobMsg{
		JobID:     "JobActionUserDownRight",
		DelayTime: 5 * 3600,
		Args:      []interface{}{7327, 8293729, 4397892},
	})
	dq.AddDelay(DelayJobMsg{
		JobID:     JobID2,
		DelayTime: 6,
		Args:      []interface{}{"job2", true, 67},
	})
	time.Sleep(20 * time.Second)
}
