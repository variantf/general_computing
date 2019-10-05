#!/bin/bash

PROTOS="proto/gc/gc-backend/general_computing.proto proto/gc/gc-backend/formula.proto proto/gc/gc-backend/data_manager.proto proto/gc/gc-backend/server.proto"
# Protocol buffer
protoc -Iproto/gc/gc-backend \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --go_out=plugins=grpc:grpc \
    --proto_path=proto/gc/gc-backend \
    $PROTOS

protoc --go_out=plugins=grpc:uma \
     --proto_path=proto/core/uma \
     proto/core/uma/uma.proto

# gRPC gateway
protoc -Iproto/gc/gc-backend \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --grpc-gateway_out=logtostderr=true:grpc \
    --proto_path=proto/gc/gc-backend \
    $PROTOS

# Swagger
protoc -Iproto/gc/gc-backend \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:grpc \
    --proto_path=proto/gc/gc-backend \
    $PROTOS
