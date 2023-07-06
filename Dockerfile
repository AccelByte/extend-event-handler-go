FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.0.0 as proto
WORKDIR /build
COPY pkg/proto pkg/proto
ENV PROTO_DIR pkg/proto/accelbyte-asyncapi
ENV PB_GO_PROTO_PATH pkg/pb/accelbyte-asyncapi
RUN mkdir -p pkg/pb
RUN for dir in $(find ${PROTO_DIR} -name '*.proto' -exec dirname {} \; | sort -u); do \
        pkg=${dir#${PROTO_DIR}/}; \
        output_dir="${PB_GO_PROTO_PATH}/$pkg"; \
        echo "Creating output directory: ${output_dir}"; \
        mkdir -p ${output_dir}; \
        echo "Compiling protobuf files in package: ${pkg}"; \
        protoc \
            --proto_path=${dir}  \
            --go_out=${output_dir} \
            --go_opt=paths=source_relative \
            --go-grpc_out=${output_dir} \
            --go-grpc_opt=paths=source_relative \
            ${dir}/*.proto; \
    done

FROM --platform=$BUILDPLATFORM golang:1.20-alpine as builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=proto /build/pkg/pb pkg/pb
RUN env GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o extend-event-handler-go_$TARGETOS-$TARGETARCH


FROM alpine:3.17.0
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
#ADD data data
COPY --from=builder /build/extend-event-handler-go_$TARGETOS-$TARGETARCH extend-event-handler-go
# Plugin arch gRPC server port
EXPOSE 6565
# Prometheus /metrics web server port
EXPOSE 8080
CMD [ "/app/extend-event-handler-go" ]
