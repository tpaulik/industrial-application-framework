// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package alarmlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

type LogType string
type AlarmSeverity string
type AlarmVisibility string

const (
	// log type
	AppAlarm   LogType = "APP_ALARM"   // for NDAC Applications
	AppFwAlarm LogType = "APPFW_ALARM" // for NDAC APPFW components

	// severity
	Minor    AlarmSeverity = "MINOR"
	Major    AlarmSeverity = "MAJOR"
	Critical AlarmSeverity = "CRITICAL"
	Warning  AlarmSeverity = "WARNING"
	Info     AlarmSeverity = "INFO"

	// visibility
	Global 	   AlarmVisibility = "GLOBAL"		// default, visible "anywhere"
	Operations AlarmVisibility = "OPERATIONS"   // not visisble in C-UI (of NDAC)
)

type AlarmDetails struct {
	Name       string `json:"name"`
	ID         string `json:"id"`
	Severity   AlarmSeverity `json:"severity"`
	Text       string `json:"text"`
	State      int    `json:"state"`
	Visibility AlarmVisibility `json:"visibility,omitempty"`
	SubDN      string `json:"subdn,omitempty"`
}

func (a *AlarmDetails) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", a.Name)
	enc.AddString("id", a.ID)
	enc.AddString("severity", string(a.Severity))
	enc.AddString("text", a.Text)
	enc.AddInt("state", a.State)

	if len(a.Visibility) > 0 {
		enc.AddString("visibility", string(a.Visibility))
	}

	if len(a.SubDN) > 0 {
		enc.AddString("subdn", a.SubDN)
	}

	return nil
}

func init() {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		DisableCaller:    true,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "ts",
			EncodeTime: zapcore.ISO8601TimeEncoder,
			LineEnding: zapcore.DefaultLineEnding,
		},
	}

	logr, err := cfg.Build()
	if nil != err {
		panic("unable to create a logger")
	}

	logger = logr
}

// RaiseAlarm produces alarm log with state 1
func RaiseAlarm(logtype LogType, alarm *AlarmDetails) {
	//check if state is correct
	if 0 == len(alarm.Visibility) {
		alarm.Visibility = Global
	}
	alarm.State = 1
	print(logtype, alarm)
}

// ClearAlarm produces alarm log with severity CLEARED and state 0
func ClearAlarm(logtype LogType, alarm *AlarmDetails) {
	//check if state is correct
	alarm.State = 0
	print(logtype, alarm)
}

func print(logtype LogType, alarm *AlarmDetails) {
	logger.Info("", zap.String("log_type", string(logtype)), zap.Object("alarm", alarm))
}

// ==============================
// InitLogger is only for testing
func InitLogger() error {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		DisableCaller:    true,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "ts",
			EncodeTime: zapcore.ISO8601TimeEncoder,
			LineEnding: zapcore.DefaultLineEnding,
		},
	}

	logr, err := cfg.Build()
	if nil != err {
		return err
	}

	logger = logr
	return nil
}
