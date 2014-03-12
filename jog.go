// Copyright 2013 Justin Wilson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jog is used to log JSON style logging information.
// It's intended to be used with Go's log package, by calling `log.SetOutput(io.Writer)`,
// although it can be used standalone as well.
package jog

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"
)

const (
	CRITICAL = Level("critical")
	ERROR    = Level("error")
	WARNING  = Level("warning")
	INFO     = Level("info")
	DEBUG    = Level("debug")
)

// Level is the level of the data being logged
type Level string

// Message is used to capture basic information to be logged.
// This message is then passed to the log function of a Logger.
type Message struct {
	Data  interface{} `json:"data"`
	Level Level       `json:"level"`
	File  string      `json:"file"`
	Line  int         `json:"line"`
	Time  time.Time   `json:"timestamp"`
}

// Logger is an interface used as the communication means for the log
type Logger interface {
	Log(m *Message) (int, error)
}

// Jog is the core logging type, it contains an instance of a Logger that is passed
// the log Message.
// Jog implements io.Writer so it can be used as log.SetOutput(logWriter)
type Jog struct {
	logger Logger
}

// Log with a given Level and object
func (j *Jog) Log(l Level, o interface{}) (int, error) {
	return j.write(newMessage(l, o, 3))
}

// Log a critical message by the given object
func (j *Jog) Critical(o interface{}) error {
	_, err := j.Log(CRITICAL, o)
	return err
}

// Log a error message by the given object
func (j *Jog) Error(o interface{}) error {
	_, err := j.Log(ERROR, o)
	return err
}

// Log a warning message by the given object
func (j *Jog) Warning(o interface{}) error {
	_, err := j.Log(WARNING, o)
	return err
}

// Log a info message by the given object
func (j *Jog) Info(o interface{}) error {
	_, err := j.Log(INFO, o)
	return err
}

// Log a debug message by the given object
func (j *Jog) Debug(o interface{}) error {
	_, err := j.Log(DEBUG, o)
	return err
}

// Writes the given bytes to a Logger
// (implementation of io.Writer)
func (j *Jog) Write(p []byte) (int, error) {
	m := newMessage(INFO, nil, 4)

	// Remove trailing "\n", added by `log.Output(int, string)`
	l := len(p) - 1
	if len(p) > 0 && p[l] == '\n' {
		p = p[0:l]
		l--
	}

	// Attempt to set JSON value of `p` and log level
	isJSONLike := l > 1 && p[0] == '{' && p[l] == '}'
	if isJSONLike && json.Unmarshal(p, &m.Data) == nil {
		m.Level = levelFrom(m.Data)
	} else {
		m.Data = string(p)
	}

	// Send to logger
	return j.write(m)
}

// Invoke the Logger with the JSON data
func (j *Jog) write(m *Message) (int, error) {
	n, err := j.logger.Log(m)
	if err != nil {
		s := fmt.Sprintf("[LOG FAILURE] - (Logger) %s -> %#v\n", err, m)
		os.Stderr.Write([]byte(fmt.Sprintf("%v", s)))
		return 0, err
	}
	return n, err
}

// Pulls the `level` value from the message to be logged
func levelFrom(o interface{}) Level {
	level := INFO
	// Ensure we've got data
	m, ok := o.(map[string]interface{})
	if !ok {
		return level
	}

	// Ensure the key `level` exists
	l, ok := m["level"]
	if !ok {
		return level
	}

	// Remove `level` from data, as it exists in the message
	delete(m, "level")

	// Set the level
	switch l {
	case "critical":
		level = CRITICAL
	case "error":
		level = ERROR
	case "warning":
		level = WARNING
	case "debug":
		level = DEBUG
	}

	return level
}

func newMessage(l Level, d interface{}, depth int) *Message {
	m := &Message{
		Data:  d,
		Level: l,
		Time:  time.Now().UTC(),
		File:  "???",
		Line:  0,
	}

	// Set filename/line number of invoker
	if _, file, line, ok := runtime.Caller(depth); ok {
		m.File, m.Line = file, line
	}

	if d == nil {
		return m
	}

	// Set Data/Message fields
	if s, ok := d.(string); ok {
		m.Data = s
	} else if b, err := json.Marshal(d); err != nil || len(b) < 3 {
		m.Data = fmt.Sprint(d)
	}

	return m
}

// NewWriter returns an io.Writer used to write custom log messages
func NewWriter(l Logger) io.Writer {
	return &Jog{l}
}

// New returns a new Logger using a Jog logger
func NewLogger(l Logger) *log.Logger {
	return log.New(&Jog{l}, "", 0)
}

// New returns a new Jog instance
func New(l Logger) *Jog {
	return &Jog{l}
}
