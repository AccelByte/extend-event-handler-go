# Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
# This is licensed software from AccelByte Inc, for limitations
# and restrictions contact your company contract manager.

SHELL := /bin/bash

PROJECT_NAME := $(shell basename "$$(pwd)")
GOLANG_IMAGE := golang:1.20

BUILD_CACHE_VOLUME := $(shell echo '$(PROJECT_NAME)' | sed 's/[^a-zA-Z0-9_-]//g')-build-cache

IMAGE_NAME := $(shell basename "$$(pwd)")-app
BUILDER := extend-builder

build: proto
	docker run -t --rm \
			-v $(BUILD_CACHE_VOLUME):/tmp/build-cache \
			$(GOLANG_IMAGE) \
			chown $$(id -u):$$(id -g) /tmp/build-cache		# For MacOS docker host: Workaround for /tmp/build-cache folder
	docker run -t --rm \
			-u $$(id -u):$$(id -g) \
			-e GOCACHE=/tmp/build-cache/go/cache \
			-e GOMODCACHE=/tmp/build-cache/go/modcache \
			-v $(BUILD_CACHE_VOLUME):/tmp/build-cache \
			-v $$(pwd):/data/ \
			-w /data/ \
			$(GOLANG_IMAGE) \
			go build -modcacherw

proto:
	docker run -t --rm -u $$(id -u):$$(id -g) \
		-v $$(pwd):/data \
		-w /data \
		--entrypoint /bin/bash \
		rvolosatovs/protoc:4.1.0 \
			proto.sh

imagex:
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${IMAGE_NAME} --platform linux/amd64 .
	docker buildx build -t ${IMAGE_NAME} --load .
	docker buildx rm --keep-state $(BUILDER)

imagex_push:
	@test -n "$(IMAGE_TAG)" || (echo "IMAGE_TAG is not set (e.g. 'v0.1.0', 'latest')"; exit 1)
	@test -n "$(REPO_URL)" || (echo "REPO_URL is not set"; exit 1)
	docker buildx inspect $(BUILDER) || docker buildx create --name $(BUILDER) --use
	docker buildx build -t ${REPO_URL}:${IMAGE_TAG} --platform linux/amd64 --push .
	docker buildx rm --keep-state $(BUILDER)
