// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package logger

import (
	"bytes"
	"fmt"
	"time"
)

const (
	infoStr = "INFO "
	errStr  = "ERROR "
)

// Simple logger interface for reporting logs to bosh
type Logger interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	LogData() string
}

type bufferLogger struct {
	buffer *bytes.Buffer
}

func New() Logger {
	return bufferLogger{&bytes.Buffer{}}
}

func (l bufferLogger) Info(v ...interface{}) {
	l.buffer.WriteString(timestamp() + infoStr + fmt.Sprint(v...) + "\n")
}

func (l bufferLogger) Infof(format string, v ...interface{}) {
	l.buffer.WriteString(timestamp() + infoStr + fmt.Sprintf(format, v...) + "\n")
}

func (l bufferLogger) Error(v ...interface{}) {
	l.buffer.WriteString(timestamp() + errStr + fmt.Sprint(v...) + "\n")
}

func (l bufferLogger) Errorf(format string, v ...interface{}) {
	l.buffer.WriteString(timestamp() + errStr + fmt.Sprintf(format, v...) + "\n")
}

func (l bufferLogger) LogData() string {
	return l.buffer.String()
}

func timestamp() string {
	// UTC time formatted as RFC3339, retains order when sorted as a string
	return time.Now().UTC().Format(time.RFC3339) + " " // separator for parsing
}
