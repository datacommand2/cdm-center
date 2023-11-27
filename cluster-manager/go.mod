module github.com/datacommand2/cdm-center/cluster-manager

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/datacommand2/cdm-cloud/common v0.0.0-20231127061122-07e02be5bd0c
	github.com/datacommand2/cdm-cloud/services/identity v0.0.0-20231127061639-e680b139acd3
	github.com/datacommand2/cdm-disaster-recovery/common v0.0.0-20231127064704-ec552445103d
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/jinzhu/copier v0.3.4
	github.com/jinzhu/gorm v1.9.16
	github.com/micro/go-micro/v2 v2.9.1
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/protobuf v1.28.1
)
