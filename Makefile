# Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

SHELL := /bin/bash

GOLANG_DOCKER_IMAGE := golang:1.20
IMAGE_NAME := $(shell basename "$$(pwd)")-app
BUILDER := grpc-plugin-server-builder
PROTO_DIR := pkg/proto
PB_GO_PROTO_PATH := pkg/pb
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
PROTO_FILES_REL := $(subst $(PROTO_DIR)/,,$(PROTO_FILES))

.PHONY: proto

proto:
	rm -rfv $(PB_GO_PROTO_PATH)/*
	for proto_file in $(PROTO_FILES_REL); do \
		proto_pkg=$$(dirname $$proto_file); \
		mkdir -p $(PB_GO_PROTO_PATH)/$$proto_pkg; \
		docker run -t --rm -u $$(id -u):$$(id -g) -v $$(pwd):/data/ -w /data/ rvolosatovs/protoc:4.0.0 \
			--proto_path=$(PROTO_DIR) \
        	--go_out=$(PB_GO_PROTO_PATH) \
        	--go_opt=paths=source_relative \
           	--go-grpc_out=$(PB_GO_PROTO_PATH) \
           	--go-grpc_opt=paths=source_relative \
           	$(PROTO_DIR)/$$proto_file; \
	done

lint: proto
	rm -f lint.err
	find -type f -iname go.mod -exec dirname {} \; | while read DIRECTORY; do \
		echo "# $$DIRECTORY"; \
		docker run -t --rm -u $$(id -u):$$(id -g) -v $$(pwd):/data/ -w /data/ -e GOCACHE=/data/.cache/go-build -e GOLANGCI_LINT_CACHE=/data/.cache/go-lint golangci/golangci-lint:v1.42.1\
				sh -c "cd $$DIRECTORY && golangci-lint -v --timeout 5m --max-same-issues 0 --max-issues-per-linter 0 --color never run || touch /data/lint.err"; \
	done
	[ ! -f lint.err ] || (rm lint.err && exit 1)

build: proto
	docker run -t --rm -u $$(id -u):$$(id -g) -v $$(pwd):/data/ -w /data/ -e GOCACHE=/data/.cache/go-build $(GOLANG_DOCKER_IMAGE) \
		sh -c "go build -buildvcs=false"

#image:
#	docker buildx build -t ${IMAGE_NAME} --load .
#

imagex:
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${IMAGE_NAME} --platform linux/arm64/v8,linux/amd64 .
	docker buildx build -t ${IMAGE_NAME} --load .
	docker buildx rm --keep-state $(BUILDER)

imagex_push:
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is not set (e.g. 'v0.1.0', 'latest')"; exit 1)
	@test -n "$(REPO_URL)" || (echo "REPO_URL is not set"; exit 1)
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${REPO_URL}:${IMAGE_TAG} --platform linux/arm64/v8,linux/amd64 --push .
	docker buildx rm --keep-state $(BUILDER)

