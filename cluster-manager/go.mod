module github.com/datacommand2/cdm-center/cluster-manager

go 1.14

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2
	github.com/datacommand2/cdm-cloud/common v0.0.0-20231127061122-07e02be5bd0c
	github.com/datacommand2/cdm-cloud/services/identity v0.0.0-20231127061639-e680b139acd3
	github.com/datacommand2/cdm-disaster-recovery/common v0.0.0-20231127071405-33c955c2dcbd
	github.com/datacommand2/cdm-disaster-recovery/manager v0.0.0-20231128001023-78919424bb16
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.4.0
	github.com/jinzhu/copier v0.4.0
	github.com/jinzhu/gorm v1.9.16
	github.com/micro/go-micro/v2 v2.9.1
	github.com/streadway/amqp v1.0.0
	golang.org/x/net v0.17.0
	google.golang.org/genproto/googleapis/api v0.0.0-20231106174013-bbf56f31fb17
	google.golang.org/protobuf v1.31.0
)
