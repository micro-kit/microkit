module github.com/micro-kit/microkit

go 1.14

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/micro-kit/micro-common => ../../micro-kit/micro-common
	go.etcd.io/bbolt => github.com/coreos/bbolt v1.3.4
	google.golang.org/grpc => google.golang.org/grpc v1.26.0 // grpc对etcd依赖问题
)

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/coreos/bbolt v1.3.2 // indirect
	github.com/coreos/etcd v3.3.20+incompatible
	github.com/golang/protobuf v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/joho/godotenv v1.3.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/micro-kit/micro-common v0.0.0-00010101000000-000000000000
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/procfs v0.0.11 // indirect
	github.com/sirupsen/logrus v1.5.0 // indirect
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/soheilhy/cmux v0.1.4
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	go.etcd.io/bbolt v1.3.2 // indirect
	go.etcd.io/etcd v3.3.20+incompatible
	go.uber.org/zap v1.14.1
	golang.org/x/crypto v0.0.0-20200414173820-0848c9571904 // indirect
	golang.org/x/sys v0.0.0-20200413165638-669c56c373c4 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/genproto v0.0.0-20200413115906-b5235f65be36 // indirect
	google.golang.org/grpc v1.28.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
