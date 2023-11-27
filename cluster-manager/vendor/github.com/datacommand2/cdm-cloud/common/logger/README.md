# Logger

CDM-Cloud 의 공통 로깅 라이브러리로써 레벨에 맞는 로그를 표준출력 혹은 로그파일에 출력하는 기능을 구현한다.
로깅 레벨은 Trace, Debug, Info, Warn, Error, Fatal 이 있으며, 각각의 레벨은 다음과 같은 경우 사용된다.

- Trace: DEBUG보다 세분화 된 정보 이벤트를 지정합니다.
- Debug: 응용 프로그램을 디버깅하는 데 유용한 세분화 된 정보 이벤트를 지정합니다.
- Info: 응용 프로그램의 진행 상황을 대략적인 수준으로 강조 표시하는 정보 메시지를 지정합니다.
- Warn: 잠재적으로 유해한 상황을 나타냅니다.
- Error: 응용 프로그램이 계속 실행될 수 있는 오류 이벤트를 지정합니다.
- Fatal: 응용 프로그램이 중단되는 심각한 오류 이벤트를 지정합니다.


## Usage
Logger 라이브러리를 사용하기 위해 다음 패키지를 import 한다.
```go
import "github.com/datacommand2/cdm-cloud/common/logger"
```

Logger 를 초기화한 뒤 사용할 수 있으며, 초기화 하지 않을 경우 기본적으로
**os.Stdout** 에 **Info** 레벨을 포함한 상위 레벨의 로그 테스트를 출력한다.
초기화 및 옵션 함수는 다음과 같으며, 추가로 go-micro 의 logger 옵션 함수를 사용할 수 있다.

```go
// Logger 를 초기화한다. 초기화하지 않을 경우 **Info** 포함 상위 레벨의 로그를 **os.Stdout** 에 출력한다.
func Init(opts ...Option)

// 서비스 이름을 설정한다.
// 서비스 이름이 설정되면 서비스 이름이 로그 텍스트의 prefix 에 붙는다.
func WithServiceName(name string) Option
```

로깅 텍스트는 `2020-06-08 05:54:42 [로깅레벨] 로그내용 (func file:line)` 포멧으로 출력되며, 로깅일시는 UTC 를 기준으로 출력된다.
로깅을 위해 제공되는 함수는 다음과 같다.

```go
// 로그 텍스트를 Trace 레벨로 기록한다.
func Trace(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Trace 레벨로 기록한다.
func Tracef(format string, v ...interface{})

// 로그 텍스트를 Debug 레벨로 기록한다.
func Debug(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Debug 레벨로 기록한다.
func Debugf(format string, v ...interface{})

// 로그 텍스트를 Info 레벨로 기록한다.
func Info(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Info 레벨로 기록한다.
func Infof(format string, v ...interface{})

// 로그 텍스트를 Warning 레벨로 기록한다.
func Warn(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Warning 레벨로 기록한다.
func Warnf(format string, v ...interface{})

// 로그 텍스트를 Error 레벨로 기록한다.
func Error(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Error 레벨로 기록한다.
func Errorf(format string, v ...interface{})

// 로그 텍스트를 Fatal 레벨로 기록하고, 프로그램을 종료한다.
func Fatal(v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 Fatal 레벨로 기록하고, 프로그램을 종료한다.
func Fatalf(format string, v ...interface{})
```

기본 Logger 외에 별도의 Logger 를 생성해서 사용할 수 있으며, Logger 를 생성하는 함수와 로깅 함수는 다음과 같다.
```go
// NewLogger 는 io.Writer 에 로깅하는 defaultLogger 를 생성하는 함수이다.
func NewLogger(opts ...Option) Logger

// 로그 텍스트를 지정한 레벨로 기록한다.
func Log(level Level, v ...interface{})

// 포맷 템플릿을 적용한 로그 텍스트를 지정한 레벨로 기록한다.
func Logf(level Level, format string, v ...interface{})
```
