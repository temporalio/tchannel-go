// Copyright (c) 2015 Uber Technologies, Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package tchannel_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/temporalio/tchannel-go"

	"github.com/stretchr/testify/assert"
)

func field(k string, v interface{}) tchannel.LogField {
	return tchannel.LogField{Key: k, Value: v}
}

func TestErrField(t *testing.T) {
	assert.Equal(t, field("error", "foo"), tchannel.ErrField(errors.New("foo")))
}

func TestWriterLogger(t *testing.T) {
	var buf bytes.Buffer
	var bufLogger = tchannel.NewLogger(&buf)

	debugf := func(logger tchannel.Logger, msg string, args ...interface{}) { logger.Debugf(msg, args...) }
	infof := func(logger tchannel.Logger, msg string, args ...interface{}) { logger.Infof(msg, args...) }

	levels := []struct {
		levelFunc   func(logger tchannel.Logger, msg string, args ...interface{})
		levelPrefix string
	}{
		{debugf, "D"},
		{infof, "I"},
	}

	for _, level := range levels {
		tagLogger1 := bufLogger.WithFields(field("key1", "value1"))
		tagLogger2 := bufLogger.WithFields(field("key2", "value2"), field("key3", "value3"))

		verifyMsgAndPrefix := func(logger tchannel.Logger) {
			buf.Reset()
			level.levelFunc(logger, "mes%v", "sage")

			out := buf.String()
			assert.Contains(t, out, "message")
			assert.Contains(t, out, "["+level.levelPrefix+"]")
		}

		verifyMsgAndPrefix(bufLogger)

		verifyMsgAndPrefix(tagLogger1)
		assert.Contains(t, buf.String(), "{key1 value1}")
		assert.NotContains(t, buf.String(), "{key2 value2}")
		assert.NotContains(t, buf.String(), "{key3 value3}")

		verifyMsgAndPrefix(tagLogger2)
		assert.Contains(t, buf.String(), "{key2 value2}")
		assert.Contains(t, buf.String(), "{key3 value3}")
		assert.NotContains(t, buf.String(), "{key1 value1}")
	}
}

func TestWriterLoggerNoSubstitution(t *testing.T) {
	var buf bytes.Buffer
	var bufLogger = tchannel.NewLogger(&buf)

	logDebug := func(logger tchannel.Logger, msg string) { logger.Debug(msg) }
	logInfo := func(logger tchannel.Logger, msg string) { logger.Info(msg) }
	logWarn := func(logger tchannel.Logger, msg string) { logger.Warn(msg) }
	logError := func(logger tchannel.Logger, msg string) { logger.Error(msg) }

	levels := []struct {
		levelFunc   func(logger tchannel.Logger, msg string)
		levelPrefix string
	}{
		{logDebug, "D"},
		{logInfo, "I"},
		{logWarn, "W"},
		{logError, "E"},
	}

	for _, level := range levels {
		tagLogger1 := bufLogger.WithFields(field("key1", "value1"))
		tagLogger2 := bufLogger.WithFields(field("key2", "value2"), field("key3", "value3"))

		verifyMsgAndPrefix := func(logger tchannel.Logger) {
			buf.Reset()
			level.levelFunc(logger, "test-msg")

			out := buf.String()
			assert.Contains(t, out, "test-msg")
			assert.Contains(t, out, "["+level.levelPrefix+"]")
		}

		verifyMsgAndPrefix(bufLogger)

		verifyMsgAndPrefix(tagLogger1)
		assert.Contains(t, buf.String(), "{key1 value1}")
		assert.NotContains(t, buf.String(), "{key2 value2}")
		assert.NotContains(t, buf.String(), "{key3 value3}")

		verifyMsgAndPrefix(tagLogger2)
		assert.Contains(t, buf.String(), "{key2 value2}")
		assert.Contains(t, buf.String(), "{key3 value3}")
		assert.NotContains(t, buf.String(), "{key1 value1}")
	}
}

func TestLevelLogger(t *testing.T) {
	var buf bytes.Buffer
	var bufLogger = tchannel.NewLogger(&buf)

	expectedLines := map[tchannel.LogLevel]int{
		tchannel.LogLevelAll:   6,
		tchannel.LogLevelDebug: 6,
		tchannel.LogLevelInfo:  4,
		tchannel.LogLevelWarn:  2,
		tchannel.LogLevelError: 1,
		tchannel.LogLevelFatal: 0,
	}
	for level := tchannel.LogLevelFatal; level >= tchannel.LogLevelAll; level-- {
		buf.Reset()
		levelLogger := tchannel.NewLevelLogger(bufLogger, level)

		for l := tchannel.LogLevel(0); l <= tchannel.LogLevelFatal; l++ {
			assert.Equal(t, level <= l, levelLogger.Enabled(l), "levelLogger.Enabled(%v) at %v", l, level)
		}

		levelLogger.Debug("debug")
		levelLogger.Debugf("debu%v", "g")
		levelLogger.Info("info")
		levelLogger.Infof("inf%v", "o")
		levelLogger.Warn("warn")
		levelLogger.Error("error")

		assert.Equal(t, expectedLines[level], bytes.Count(buf.Bytes(), []byte{'\n'}))
	}
}
