// (c) Copyright IBM Corp. 2022

//go:build go1.16
// +build go1.16

package instaredigo

import (
	"context"
	"errors"
	"testing"
	"time"

	instana "github.com/instana/go-sensor"
	"github.com/instana/testify/assert"
	"github.com/instana/testify/require"
)

type MockConn struct {
	address string
}

func (conn *MockConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	reply = "OK <->" + commandName
	if len(commandName) == 0 {
		err = errors.New("Empty command received")
	}
	return reply, err
}

func (conn *MockConn) DoContext(ctx context.Context,commandName string, 
    args ...interface{}) (reply interface{}, err error) {
	reply = "OK <->" + commandName
	if len(commandName) == 0 {
		err = errors.New("Empty command received")
	}
	return reply, err
}

func (conn *MockConn) ReceiveContext(ctx context.Context) (reply interface{}, err error) {
    reply = "OK"
    return reply, err
}

func (conn *MockConn) DoWithTimeout(timeOut time.Duration, commandName string, 
    args ...interface{}) (reply interface{}, err error) {
	reply = "OK <->" + commandName
	if len(commandName) == 0 {
		err = errors.New("Empty command received")
	}
	return reply, err
}

func (conn *MockConn) ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error) {
    reply = "OK"
    return reply, err
}

func (conn *MockConn) Send(commandName string, args ...interface{}) error {
	var err error
	if len(commandName) == 0 {
		err = errors.New("Empty command received")
	}
	return err
}

func (conn *MockConn) Receive() (reply interface{}, err error) {
	reply = "OK"
	return reply, err
}

func (conn *MockConn) Err() error {
	err := errors.New("No error")
	return err
}

func (conn *MockConn) Close() error {
	err := errors.New("No error")
	return err
}

func (conn *MockConn) Flush() error {
	err := errors.New("No error")
	return err
}

func TestMockDo(t *testing.T) {
	examples := map[string]struct {
		DoCommand []interface{}
		Expected  instana.RedisSpanTags
	}{
		"SET": {
			DoCommand: []interface{}{"name", "Instana"},
			Expected: instana.RedisSpanTags{
				Command: "SET",
			},
		},
		"GET": {
			DoCommand: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "GET",
			},
		},
		"DEL": {
			DoCommand: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "DEL",
			},
		},
	}
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)
			sp := sensor.Tracer().StartSpan("testing")
			defer sp.Finish()
			conn := &instaRedigoConn{&MockConn{}, sensor, ":7001", nil}
            defer conn.Close()
			_, err := conn.Do(name, example.DoCommand...)
			assert.Equal(t, err, nil)
			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			dbSpan := spans[0]
			data := dbSpan.Data.(instana.RedisSpanData)

			assert.Equal(t, "redis", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

			assert.Equal(t, example.Expected.Error, data.Tags.Error)
			assert.Equal(t, example.Expected.Command, data.Tags.Command)
		})
	}
}

func TestMockSend(t *testing.T) {
	examples := map[string]struct {
		DoCommand []interface{}
		Expected  instana.RedisSpanTags
	}{
		"SET": {
			DoCommand: []interface{}{"name", "Instana"},
			Expected: instana.RedisSpanTags{
				Command: "SET",
			},
		},
		"GET": {
			DoCommand: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "GET",
			},
		},
		"DEL": {
			DoCommand: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "DEL",
			},
		},
	}
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)
			sp := sensor.Tracer().StartSpan("testing")
			defer sp.Finish()
			conn := &instaRedigoConn{&MockConn{}, sensor, ":7001", nil}
            defer conn.Close()
			err := conn.Send(name, example.DoCommand...)
			assert.Equal(t, err, nil)
			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			dbSpan := spans[0]
			data := dbSpan.Data.(instana.RedisSpanData)

			assert.Equal(t, "redis", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

			assert.Equal(t, example.Expected.Error, data.Tags.Error)
			assert.Equal(t, example.Expected.Command, data.Tags.Command)
		})
	}
}


func TestSubCommands(t *testing.T) {
	testCases := map[string]struct {
		BatchCommands [][]interface{}
		Expected      instana.RedisSpanTags
	}{
		"batch commands": {
			BatchCommands: [][]interface{}{
				{"multi"},
				{"set", "name", "IBM"},
				{"get", "name"},
				{"del", "name"},
				{"exec"},
			},
			Expected: instana.RedisSpanTags{
				Command:     "multi",
				Subcommands: []string{"set", "get", "del"},
			},
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)
			sp := sensor.Tracer().StartSpan("testing")
			defer sp.Finish()
			conn := &instaRedigoConn{&MockConn{}, sensor, ":7001", nil}
			defer conn.Close()
			ctx := context.Background()
			ctxSpan := instana.ContextWithSpan(ctx, sp)
			for _, cmd := range testCase.BatchCommands {
				cmdArgs := cmd[1:]
				cmdArgs = append(cmdArgs, ctxSpan)
				cmdStr := cmd[0].(string)
                err := conn.Send(cmdStr, cmdArgs...)
				assert.Equal(t, nil, err)
			}
			spans := recorder.GetQueuedSpans()
			dbSpan := spans[0]
			assert.Equal(t, 1, len(spans))
			data := dbSpan.Data.(instana.RedisSpanData)
			assert.Equal(t, "redis", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

			assert.Equal(t, testCase.Expected.Error, data.Tags.Error)
			assert.Equal(t, testCase.Expected.Command, data.Tags.Command)
			assert.Equal(t, testCase.Expected.Subcommands, data.Tags.Subcommands)
		})
	}
}

func TestMockDoContext(t *testing.T) {
	examples := map[string]struct {
		Command []interface{}
		Expected  instana.RedisSpanTags
	}{
		"SET": {
			Command: []interface{}{"name", "Instana"},
			Expected: instana.RedisSpanTags{
				Command: "SET",
			},
		},
		"GET": {
			Command: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "GET",
			},
		},
		"DEL": {
			Command: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "DEL",
			},
		},
	}
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)
			sp := sensor.Tracer().StartSpan("testing")
			defer sp.Finish()
			conn := &instaRedigoConn{&MockConn{}, sensor, ":7001", nil}
            defer conn.Close()
            ctx := context.Background()
            _, err := conn.DoContext(ctx, name, example.Command...)
			assert.Equal(t, err, nil)
			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			dbSpan := spans[0]
			data := dbSpan.Data.(instana.RedisSpanData)

			assert.Equal(t, "redis", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

			assert.Equal(t, example.Expected.Error, data.Tags.Error)
			assert.Equal(t, example.Expected.Command, data.Tags.Command)
		})
	}
}

func TestMockDoTimeout(t *testing.T) {
	examples := map[string]struct {
		Command []interface{}
		Expected  instana.RedisSpanTags
	}{
		"SET": {
			Command: []interface{}{"name", "Instana"},
			Expected: instana.RedisSpanTags{
				Command: "SET",
			},
		},
		"GET": {
			Command: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "GET",
			},
		},
		"DEL": {
			Command: []interface{}{"name"},
			Expected: instana.RedisSpanTags{
				Command: "DEL",
			},
		},
	}
	for name, example := range examples {
		t.Run(name, func(t *testing.T) {
			recorder := instana.NewTestRecorder()
			sensor := instana.NewSensorWithTracer(
				instana.NewTracerWithEverything(instana.DefaultOptions(), recorder),
			)
			sp := sensor.Tracer().StartSpan("testing")
			defer sp.Finish()
			conn := &instaRedigoConn{&MockConn{}, sensor, ":7001", nil}
            defer conn.Close()
            _, err := conn.DoWithTimeout(1000, name, example.Command...)
			assert.Equal(t, err, nil)
			spans := recorder.GetQueuedSpans()
			assert.Equal(t, 1, len(spans))
			dbSpan := spans[0]
			data := dbSpan.Data.(instana.RedisSpanData)

			assert.Equal(t, "redis", dbSpan.Name)
			assert.EqualValues(t, instana.ExitSpanKind, dbSpan.Kind)
			assert.Empty(t, dbSpan.Ec)

			require.IsType(t, instana.RedisSpanData{}, dbSpan.Data)

			assert.Equal(t, example.Expected.Error, data.Tags.Error)
			assert.Equal(t, example.Expected.Command, data.Tags.Command)
		})
	}
}
