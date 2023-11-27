package logger

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/micro/go-micro/v2/logger"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger is a generic logging interface
type defaultLogger struct {
	lock sync.Mutex
	opts logger.Options
}

const basePath = "/var/log"

// Init 은 Logger 를 초기화하는 함수이다.
func (l *defaultLogger) Init(opts ...logger.Option) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	for _, o := range opts {
		o(&l.opts)
	}

	return nil
}

// Options 는 Logger 의 옵션을 반환하는 함수이다.
func (l *defaultLogger) Options() logger.Options {
	l.lock.Lock()
	defer l.lock.Unlock()

	opts := logger.Options{}
	_ = copier.Copy(&opts, l.opts)

	return opts
}

// Fields 는 사용하지 않는 함수이다.
func (l *defaultLogger) Fields(_ map[string]interface{}) logger.Logger {
	return l
}

// Log 는 로그 텍스트를 기록하는 함수이다.
func (l *defaultLogger) Log(level logger.Level, v ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 옵션에 설정된 레벨 혹은 상위 레벨의 로그 텍스트만 기록한다.
	if !l.opts.Level.Enabled(level) {
		return
	}

	var text = []string{
		fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%-5s]", strings.ToUpper(level.String())),
		fmt.Sprintf("%-40s", fmt.Sprint(v...)),
	}

	// 로깅함수를 호출한 함수명, 코드파일명, 라인번호를 같이 로깅한다.
	if c, _, _, ok := runtime.Caller(l.opts.CallerSkipCount); ok {
		f, _ := runtime.CallersFrames([]uintptr{c}).Next()
		text = append(text, fmt.Sprintf("(%s %s:%d)", path.Base(f.Function), path.Base(f.File), f.Line))
	}

	_, _ = fmt.Fprintln(l.opts.Out, strings.Join(text, " "))
}

// Logf 는 포맷 템플릿을 적용한 로그 텍스트를 기록하는 함수이다.
func (l *defaultLogger) Logf(level logger.Level, format string, v ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 옵션에 설정된 레벨 혹은 상위 레벨의 로그 텍스트만 기록한다.
	if !l.opts.Level.Enabled(level) {
		return
	}

	var text = []string{
		fmt.Sprintf("%s", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("[%-5s]", strings.ToUpper(level.String())),
		fmt.Sprintf("%-40s", fmt.Sprintf(format, v...)),
	}

	// 로깅함수를 호출한 함수명, 코드파일명, 라인번호를 같이 로깅한다.
	if c, _, _, ok := runtime.Caller(l.opts.CallerSkipCount); ok {
		f, _ := runtime.CallersFrames([]uintptr{c}).Next()
		text = append(text, fmt.Sprintf("(%s %s:%d)", path.Base(f.Function), path.Base(f.File), f.Line))
	}

	_, _ = fmt.Fprintln(l.opts.Out, strings.Join(text, " "))
}

// String 은 사용하지 않는 함수이다.
func (l *defaultLogger) String() string {
	return ""
}

// NewLogger 는 Logger 를 생성하는 함수이다.
// 옵션을 지정하지 않을 경우, io.Writer 에 INFO 레벨의 로그 텍스트를 기록하는 Logger 가 생성된다.
func NewLogger(opts ...logger.Option) logger.Logger {
	hostname, _ := os.Hostname()
	logf, err := rotatelogs.New(
		"/var/log/"+hostname+"-%Y-%m-%d.log",
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(24*time.Hour*30),
		// 사이즈 제한: 50000000Byte = 47.68MB. 사이즈가 넘으면 [파일이름.log.0n] 으로 남음
		rotatelogs.WithRotationSize(50000000),
	)
	if err != nil {
		logger.Warnf("Unknown error : %v", err)
	}

	options := logger.Options{
		Out:             io.MultiWriter(logf, os.Stdout),
		Level:           logger.InfoLevel,
		CallerSkipCount: 2,
		Context:         context.Background(),
	}

	l := &defaultLogger{opts: options}
	_ = l.Init(opts...)

	return l
}
