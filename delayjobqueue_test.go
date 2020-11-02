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

type JobAction1 struct{}

func (JobAction1) ID() string {
	return JobID1
}

func (JobAction1) Execute(args []interface{}) error {
	fmt.Printf("This is %s, args=%v\n", JobID1, args)
	return nil
}

type JobAction2 struct{}

func (JobAction2) ID() string {
	return JobID1
}

func (JobAction2) Execute(args []interface{}) error {
	fmt.Printf("This is %s, args=%v\n", JobID2, args)
	return nil
}

func TestNew(t *testing.T) {
	dq := New(DelayQConf{Redis: RedisConf{
		Host: "127.0.0.1",
		Port: 6379,
	}})
	dq.Register(JobAction1{})
	dq.Register(JobAction2{})
	dq.StartBackground()
	time.Sleep(2 * time.Second)
	fmt.Println("add delay")
	dq.AddDelay(DelayJobMsg{
		JobID:     JobID1,
		DelayTime: 2,
		Args:      []interface{}{1, 2, 3, "dasd", true},
	})
	dq.AddDelay(DelayJobMsg{
		JobID:     JobID2,
		DelayTime: 1,
		Args:      []interface{}{"job2", 34, 1},
	})
	dq.AddDelay(DelayJobMsg{
		JobID:     JobID2,
		DelayTime: 6,
		Args:      []interface{}{"job2", true, 67},
	})
	time.Sleep(20 * time.Second)
}
