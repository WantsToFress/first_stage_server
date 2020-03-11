#!/usr/bin/env bash
# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
# go get -u github.com/golang/protobuf/protoc-gen-go
protoc -I api/proto --grpc-gateway_out=logtostderr=true:pkg --go_out=plugins=grpc:pkg --swagger_out=logtostderr=true:api/swagger-ui api/proto/event.proto
