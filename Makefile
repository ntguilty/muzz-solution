generate-proto:
	protoc --go_out=$(shell pwd)/pkg/proto --go-grpc_out=$(shell pwd)/pkg/proto internal/explore/adapters/grpc/explore.proto
