module github.com/voc/rtmp-auth

go 1.17

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/csrf v1.6.2
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/consul/api v1.11.0
	github.com/pelletier/go-toml v1.9.4
	github.com/rakyll/statik v0.1.7
	google.golang.org/protobuf v1.30.0
)
