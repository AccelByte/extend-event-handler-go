# Copyright (c) 2023 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

SHELL := /bin/bash

GOLANG_DOCKER_IMAGE := golang:1.20
IMAGE_NAME := $(shell basename "$$(pwd)")-app
BUILDER := grpc-plugin-server-builder
PROTO_DIR := pkg/proto/accelbyte-asyncapi
PB_GO_PROTO_PATH := pkg/pb/accelbyte-asyncapi
PROTO_SUB_DIRS := $(shell find $(PROTO_DIR) -name '*.proto' -printf '%h\n')
PROTO_PACKAGES := $(subst $(PROTO_DIR)/,,$(PROTO_SUB_DIRS))
PROTO_PACKAGES := $(sort $(PROTO_PACKAGES))


.PHONY: proto

proto:
	rm -rfv pkg/pb/*
	for pkg in $(PROTO_PACKAGES); do \
		mkdir -p $(PB_GO_PROTO_PATH)/$$pkg; \
		docker run -t --rm -u $$(id -u):$$(id -g) -v $$(pwd):/data/ -w /data/ rvolosatovs/protoc:4.0.0 \
			--proto_path=$(PROTO_DIR)/$$pkg  \
			--go_out=$(PB_GO_PROTO_PATH)/$$pkg \
			--go_opt=paths=source_relative \
   			--go-grpc_out=$(PB_GO_PROTO_PATH)/$$pkg \
   			--go-grpc_opt=paths=source_relative \
   			$(PROTO_DIR)/$$pkg/*.proto; \
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
		sh -c "go build"

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

