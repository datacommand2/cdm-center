module github.com/datacommand2/cdm-center/cluster-manager

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/datacommand2/cdm-cloud/common v0.0.0-20231205042820-988cc8ba20e0
	github.com/datacommand2/cdm-cloud/services/identity v0.0.0-20231205055758-216c42add282
	github.com/datacommand2/cdm-disaster-recovery/common v0.0.0-20231129233138-279c03db8ad2
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.4.0
	github.com/jinzhu/copier v0.3.4
	github.com/jinzhu/gorm v1.9.16
	github.com/micro/go-micro/v2 v2.9.1
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/protobuf v1.28.1
)
