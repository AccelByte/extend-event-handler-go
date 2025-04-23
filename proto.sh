#!/bin/bash

rm -rf pkg/pb/*
mkdir -p pkg/pb
# Generate the protobuf
find pkg/proto -name '*.proto' | while read PROTO_FILE; do
    protoc-wrapper -I/usr/include \
            --proto_path=pkg/proto \
            --go_out=pkg/pb \
            --go_opt=paths=source_relative \
            --go-grpc_out=pkg/pb \
            --go-grpc_opt=paths=source_relative \
            $PROTO_FILE
done
