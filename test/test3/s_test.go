package test1

import (
	"testing"
	"time"
)

func Test1(t *testing.T) {
	time.Sleep(1 * time.Second)
}

func Test2(t *testing.T) {
	time.Sleep(1 * time.Second)
}

func Test3(t *testing.T) {
	time.Sleep(1 * time.Second)
	t.Error("some error")
}
