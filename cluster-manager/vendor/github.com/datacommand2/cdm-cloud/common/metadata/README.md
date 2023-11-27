# metadata

CDM-Cloud 에서 Context의 정보를 가져오는 기능을 제공한다.

## Usage

Metadata 라이브러리를 사용하기 위해서는 다음 패키지를 Import 한다.

```go
import "github.com/datacommand2/cdm-cloud/common/metadata"
```

### 제공 함수

```go
// GetAuthenticatedSession 세션 정보를 가져온다.
func GetAuthenticatedSession(ctx context.Context) (string, error)

// GetAuthenticatedUser 사용자 정보를 가져온다.
func GetAuthenticatedUser(ctx context.Context) (*identity.User, error)

// GetClientIP IP 정보를 가져온다.
func GetClientIP(ctx context.Context) (string, error)

// GetTenantID 테넌트 ID 정보를 가져온다.
func GetTenantID(ctx context.Context) (uint64, error)
```

