package fengchaogo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
)

const (
	// StreamStartEvent start event 生成启动的事件
	StreamStartEvent = "start"
	// StreamAddEvent add event 生成数据的事件
	StreamAddEvent = "add"
	// StreamFinishEvent finish event 生成结束的事件
	StreamFinishEvent = "stop"
	// StreamErrorEvent error event 生成错误的事件
	StreamErrorEvent = "error"
)

// event bytes prefix 用于判断事件数据行的类型
var (
	// StartEventPrefix  start event 生成启动的事件的字节前缀
	StartEventPrefix = []byte("event: start")
	// AddEventPrefix  add event 生成数据的事件的字节前缀
	AddEventPrefix = []byte("event: add")
	// FinishEventPrefix  finish event 生成结束的事件的字节前缀
	FinishEventPrefix = []byte("event: stop")
	// ErrorEventPrefix  error event 生成错误的事件的字节前缀
	ErrorEventPrefix = []byte("event: error")
	// DataPrefix  data event 生成数据的事件的字节前缀
	DataPrefix = []byte("data: ")
)

// switchEventType 用于判断事件数据行的类型
func switchEventType(line []byte) string {
	if bytes.HasPrefix(line, StartEventPrefix) {
		return StreamStartEvent
	} else if bytes.HasPrefix(line, AddEventPrefix) {
		return StreamAddEvent
	} else if bytes.HasPrefix(line, FinishEventPrefix) {
		return StreamFinishEvent
	} else if bytes.HasPrefix(line, ErrorEventPrefix) {
		return StreamErrorEvent
	}
	return ""
}

// StreamAble 用于判断事件数据行的类型
type StreamAble interface {
	ChatCompletionResult
}

// StreamReader 流式数据读取器, 可用用来读取生成内容的数据包
type StreamReader[T StreamAble] interface {
	Read() (T, error)
}

// JsonStreamReader Json流式数据读取器
type JsonStreamReader[T StreamAble] struct {
	reader *bufio.Reader  // reader 用于读取数据
	resp   *http.Response // resp 用于关闭resp.Body

	errorHandler func(T) error // 处理错误
}

// Read 读取数据直到获得一个完整的数据包, 或者遇到错误或者遇到结束事件(包括EOF), 但一般情况不会遇到EOF
// 需要自定义处理数据流可以使用这个方法, 一般使用Stream方法, 可以更轻松的处理数据流
func (j *JsonStreamReader[T]) Read() (*T, bool, error) {
	var tmpbuffer = make([]byte, 0)
	isFinished := false
	catchError := false
	for {
		line, isPrefix, err := j.reader.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				isFinished = true
				// 已经结束的不会再有数据了, 也不会报错
				return nil, isFinished, nil
			}
			return nil, isFinished, err
		}

		if isPrefix {
			// line too long, read next buffer
			tmpbuffer = append(tmpbuffer, line...)
			continue
		}

		if len(tmpbuffer) > 0 {
			// if tmp is not empty, append it to line
			line = append(line, tmpbuffer...)
			tmpbuffer = tmpbuffer[:0]
		}

		if line == nil {
			// line is nil, read next buffer
			continue
		}

		// 判断事件
		switch switchEventType(line) {
		case StreamStartEvent:
			continue
		case StreamFinishEvent:
			// 准备结束, 下一次读取会进行收尾
			isFinished = true
			continue
		case StreamErrorEvent:
			// 准备捕获错误, 下一次读取会进行错误处理
			catchError = true
			continue
		case StreamAddEvent:
			continue
		}

		// 判断是否是可解析的数据包
		if !bytes.HasPrefix(line, DataPrefix) {
			if catchError {
				return nil, isFinished, fmt.Errorf("unhandled error event with line: %s", line)
			}
			if isFinished {
				return nil, isFinished, fmt.Errorf("unhandled finish event with line: %s", line)
			}
			continue
		}

		line = line[len(DataPrefix):]

		// 解析数据
		var msg T
		if err := json.Unmarshal(line, &msg); err != nil {
			return nil, isFinished, fmt.Errorf("failed to unmarshal stream message: %w", err)
		}

		// 处理错误
		if catchError {
			if err := j.errorHandler(msg); err != nil {
				return nil, isFinished, err
			}
			return &msg, isFinished, fmt.Errorf("unhandled error event")
		}
		// 如果没有错误处理, 并且没有遇到结束事件, 那么就是正常返回
		// 如果遇到了结束事件, 那么就是结束了
		return &msg, isFinished, nil
	}
}

// Stream 返回一个生成器函数
func (j *JsonStreamReader[T]) Stream() iter.Seq[T] {
	return func(yield func(T) bool) {
		defer j.Close()
		for {
			msg, finished, err := j.Read()
			if finished {
				return
			}
			if err != nil {
				panic(err)
			}
			if !yield(*msg) {
				return
			}
		}
	}
}

// Close 关闭数据流
func (j *JsonStreamReader[T]) Close() error {
	return j.resp.Body.Close()
}
