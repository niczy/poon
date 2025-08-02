module github.com/nic/poon/poon-cli

go 1.23.0

toolchain go1.23.3

require (
	github.com/nic/poon v0.0.0
	github.com/spf13/cobra v1.7.0
	google.golang.org/grpc v1.74.2
)

replace github.com/nic/poon => ../

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
