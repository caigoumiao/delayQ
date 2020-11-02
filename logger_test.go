package delayQ

import (
	"fmt"
	"testing"
)

var log1 = defaultLogger

func TestPrintfLogger_InfoF(t *testing.T) {
	log1.InfoF("test log name=%s, info=%s", "joe", "hello world!")
	log1.InfoF("test log name=%s, check=%t", "joe", true)
	log1.InfoF("test log name=%s, check=%v", "joe", []int{1, 2, 3, 4})
}

func TestPrintfLogger_ErrorF(t *testing.T) {
	log1.ErrorF("test errorf title=%s, num=%d", "err1", 4343242)
	log1.ErrorF("test errorf err=%v", fmt.Errorf("nil pointer exception!"))
}
