package v1

import (
	"fmt"
)

type Logger interface {
	Debug(string, interface{})
	Info(string, interface{})
	Warn(string, interface{})
	Error(string, interface{})
	Panic(string, interface{})
}

type NoLogger struct {}

func NewNopLogger() Logger {
	return &NoLogger{}
}

func (l NoLogger) Debug(module string, data interface{}) {
	fmt.Println(data)
}

func (l NoLogger) Info(module string, data interface{}) {
	fmt.Println(data)
}

func (l NoLogger) Warn(module string, data interface{}) {
	fmt.Println(data)
}

func (l NoLogger) Error(module string, data interface{}) {
	fmt.Println(data)
}

func (l NoLogger) Panic(module string, data interface{}) {
	fmt.Println(data)
}