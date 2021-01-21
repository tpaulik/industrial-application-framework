// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package alarmlogger_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	alog "github.com/nokia/industrial-application-framework/alarmlogger"
)

type completeLog struct {
	Ts      string             `json:"ts"`
	LogType alog.LogType       `json:"log_type"`
	Alarm   *alog.AlarmDetails `json:"alarm"`
}

func validateLogs(actual []byte, expected *completeLog) bool {
	var actualLog completeLog
	json.Unmarshal(actual, &actualLog)

	//don't care timestamp
	expected.Ts = actualLog.Ts

	if !reflect.DeepEqual(actualLog, *expected) {
		println("Expected response was :" + toJSONString(expected) + "Actual response is :" + toJSONString(actualLog))
		return false
	}
	return true
}

func toJSONString(payload interface{}) string {
	js, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(js)

}

func TestRaiseAlarm(t *testing.T) {
	osStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	expectedLog := &completeLog{
		LogType: alog.AppAlarm,
		Alarm: &alog.AlarmDetails{
			State: 1,
			Visibility: alog.Global,
		},
	}

	alog.InitLogger()
	alog.RaiseAlarm(alog.AppAlarm, &alog.AlarmDetails{})
	w.Close()

	alarmLog, _ := ioutil.ReadAll(r)
	os.Stderr = osStderr

	println("alarmlog = " + string(alarmLog))

	if !validateLogs(alarmLog, expectedLog) {
		t.Errorf("TestRaiseAlarm failed")
	}
}

func TestRaiseAlarm2(t *testing.T) {
	osStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	expectedLog := &completeLog{
		LogType: alog.AppAlarm,
		Alarm: &alog.AlarmDetails{
			Name: "AppNotRunning",
			ID: "1",
			Severity: alog.Warning,
			Text: "any",
			State: 1,
			Visibility: alog.Operations,
			SubDN: "/MODULE-servicesubmodule",
		},
	}

	alog.InitLogger()
	alog.RaiseAlarm(alog.AppAlarm, &alog.AlarmDetails{Name: "AppNotRunning",
		ID: "1",
		Text: "any",
		Severity: alog.Warning,
		State: 1,
		Visibility: alog.Operations,
		SubDN: "/MODULE-servicesubmodule",
	})
	w.Close()

	alarmLog, _ := ioutil.ReadAll(r)
	os.Stderr = osStderr

	println("alarmlog = " + string(alarmLog))

	if !validateLogs(alarmLog, expectedLog) {
		t.Errorf("TestRaiseAlarm failed")
	}
}

func TestClearAlarm(t *testing.T) {
	osStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	expectedLog := &completeLog{
		LogType: alog.AppAlarm,
		Alarm: &alog.AlarmDetails{
			State:    0,
		},
	}

	alog.InitLogger()
	alog.ClearAlarm(alog.AppAlarm, &alog.AlarmDetails{})
	w.Close()

	alarmLog, _ := ioutil.ReadAll(r)
	os.Stderr = osStderr

	println("alarmlog = " + string(alarmLog))

	if !validateLogs(alarmLog, expectedLog) {
		t.Errorf("TestClearAlarm failed")
	}
}

func TestClearAlarm2(t *testing.T) {
	osStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	expectedLog := &completeLog{
		LogType: alog.AppAlarm,
		Alarm: &alog.AlarmDetails{
			Name: "AppNotRunning",
			ID: "1",
			Severity: alog.Warning,
			Text: "any",
			State: 0,
			Visibility: alog.Operations,
			SubDN: "/MODULE-servicesubmodule",
		},
	}

	alog.InitLogger()
	alog.ClearAlarm(alog.AppAlarm, &alog.AlarmDetails{Name: "AppNotRunning",
		ID: "1",
		Text: "any",
		Severity: alog.Warning,
		State: 0,
		Visibility: alog.Operations,
		SubDN: "/MODULE-servicesubmodule",
	})
	w.Close()

	alarmLog, _ := ioutil.ReadAll(r)
	os.Stderr = osStderr

	println("alarmlog = " + string(alarmLog))

	if !validateLogs(alarmLog, expectedLog) {
		t.Errorf("TestClearAlarm failed")
	}
}
