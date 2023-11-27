package logger

import (
	"fmt"
	"github.com/micro/go-micro/v2/logger"
)

var (
	// DefaultLogger 는 logger 패키지의 기본 Logger 이며, default 로 io.Writer 에 로깅하는 Logger 를 사용한다.
	DefaultLogger = NewLogger()
)

// Init 은 추가 옵션을 지정하여 DefaultLogger 를 초기화하는 함수이다.
func Init(options ...logger.Option) error {
	return DefaultLogger.Init(options...)
}

// Trace 는 DefaultLogger 를 통해 로그 텍스트를 Trace 레벨로 기록하는 함수이다.
func Trace(v ...interface{}) {
	DefaultLogger.Log(logger.TraceLevel, v...)
}

// Tracef 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Trace 레벨로 기록하는 함수이다.
func Tracef(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.TraceLevel, format, v...)
}

// Debug 는 DefaultLogger 를 통해 로그 텍스트를 Debug 레벨로 기록하는 함수이다.
func Debug(v ...interface{}) {
	DefaultLogger.Log(logger.DebugLevel, v...)
}

// Debugf 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Debug 레벨로 기록하는 함수이다.
func Debugf(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.DebugLevel, format, v...)
}

// Info 는 DefaultLogger 를 통해 로그 텍스트를 Info 레벨로 기록하는 함수이다.
func Info(v ...interface{}) {
	DefaultLogger.Log(logger.InfoLevel, v...)
}

// Infof 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Info 레벨로 기록하는 함수이다.
func Infof(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.InfoLevel, format, v...)
}

// Warn 은 DefaultLogger 를 통해 로그 텍스트를 Warning 레벨로 기록하는 함수이다.
func Warn(v ...interface{}) {
	DefaultLogger.Log(logger.WarnLevel, v...)
}

// Warnf 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Warning 레벨로 기록하는 함수이다.
func Warnf(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.WarnLevel, format, v...)
}

// Error 는 DefaultLogger 를 통해 로그 텍스트를 Error 레벨로 기록하는 함수이다.
func Error(v ...interface{}) {
	DefaultLogger.Log(logger.ErrorLevel, v...)
}

// Errorf 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Error 레벨로 기록하는 함수이다.
func Errorf(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.ErrorLevel, format, v...)
}

// Fatal 은 DefaultLogger 를 통해 로그 텍스트를 Fatal 레벨로 기록하고, 프로그램을 종료하는 함수이다.
func Fatal(v ...interface{}) {
	DefaultLogger.Log(logger.FatalLevel, v...)
	panic(fmt.Sprint(v...))
}

// Fatalf 는 DefaultLogger 를 통해 포맷 템플릿을 적용한 로그 텍스트를 Fatal 레벨로 기록하고, 프로그램을 종료하는 함수이다.
func Fatalf(format string, v ...interface{}) {
	DefaultLogger.Logf(logger.FatalLevel, format, v...)
	panic(fmt.Sprintf(format, v...))
}
