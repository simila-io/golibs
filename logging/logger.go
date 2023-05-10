// Copyright 2023 The acquirecloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package logging

type (
	// Logger interface exposes some methods for application logging
	Logger interface {
		// Warnf is a function for printing Warn-level messages from the source code
		Warnf(format string, args ...interface{})
		// Infof is a function for printing Info-level messages from the source code
		Infof(format string, args ...interface{})
		// Debugf is a function for printing Debug-level messages from the source code
		Debugf(format string, args ...interface{})
		// Tracef is a function for pretty printing Trace-level messages from the source code
		Tracef(format string, args ...interface{})
		// Errorf is a function for pretty printing Error-level messages from the source code
		Errorf(format string, args ...interface{})
	}

	NewLoggerF func(loggerName string) Logger
	SetLevelF  func(lvl Level)

	// Level is one of ERROR, WARN, INFO, DEBUG, of TRACE
	Level int
)

const (
	ERROR = iota
	WARN
	INFO
	DEBUG
	TRACE
)

// NewLogger returns the new instance of Logger for the caller name. If you need some specific
// implementation, implement the Logger interface and assign the variable to it.
var NewLogger NewLoggerF = stdNewLogger

// SetLevel allows to set the logging level
var SetLevel SetLevelF = stdSetLevel
