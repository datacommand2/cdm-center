# GORM Model

이 패키지에는 GORM 에서 사용할 수 있는 모델을 정의하며, 정의된 모델은 다음과 같다.
> [database.ConnectionWrapper](../README.md) 를 사용할 수도 있다.

| 모델 | 설명 |
|---|---|
| Tenant | 테넌트 |
| TenantSolution | 테넌트가 사용할 수 있는 솔루션  |
| TenantReceiveEvent | 테넌트의 사용자가 기본적으로 수신하는 이벤트 |
| Role | 역할 |
| Group | 사용자 그룹 |
| User | 사용자 |
| UserRole | 사용자의 역할 |
| UserGroup | 사용자의 그룹 |
| UserReceiveEvent | 사용자가 수신하는 이벤트 |
| Event | 이벤트 |
| EventCode | 이벤트 코드 정의 |
| Schedule | 스케쥴 |
| Backup | 백업 |
| GlobalConfig | 전역 설정 |
| TenantConfig | 테넌스 설정 |
| ServiceConfig | 서비스 설정 |

모델을 사용하여 데이터베이스로 부터 레코드를 읽거나 쓸 수 있으며, 사용하는 방법은 https://gorm.io/docs/ 를 참조하면 된다.
- [레코드 생성](https://gorm.io/docs/create.html)
- [레코드 조회](https://gorm.io/docs/query.html)
- [레코드 수정](https://gorm.io/docs/update.html)
- [레코드 삭제](https://gorm.io/docs/delete.html)

추가로 관계 테이블을 조회하기 위해 제공되는 함수는 다음과 같다.

```go
// Groups 는 사용자의 그룹 목록을 조회하는 함수이다.
func (u *User) Groups(db *gorm.DB) ([]Group, error)

// Roles 는 사용자의 솔루션 역할 목록을 조회하는 함수이다.
func (u *User) Roles(db *gorm.DB) ([]Role, error)

// EventCode 는 이벤트의 이벤트 코드를 조회하는 함수이다.
func (e *Event) EventCode(db *gorm.DB) (*EventCode, error)

// Tenant 는 이벤트의 테넌트를 조회하는 함수이다.
func (e *Event) Tenant(db *gorm.DB) (*Tenant, error)
```
