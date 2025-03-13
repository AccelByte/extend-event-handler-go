# gRPC gen

FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.0.0 AS grpc-gen
WORKDIR /build
COPY pkg/proto pkg/proto
ENV PROTO_DIR pkg/proto
ENV PB_GO_PROTO_PATH pkg/pb
RUN mkdir -p pkg/pb
RUN find ${PROTO_DIR} -name '*.proto' -print0 | xargs -0 -n1 -I{} dirname {} | sort -u | \
    while read -r dir; do \
        pkg=${dir#${PROTO_DIR}/}; \
        output_dir="${PB_GO_PROTO_PATH}/$pkg"; \
        echo "Creating output directory: ${output_dir}"; \
        mkdir -p ${output_dir}; \
        echo "Compiling protobuf files in package: ${pkg}"; \
        protoc \
            --proto_path=${PROTO_DIR}  \
            --go_out=${PB_GO_PROTO_PATH} \
            --go_opt=paths=source_relative \
            --go-grpc_out=${PB_GO_PROTO_PATH} \
            --go-grpc_opt=paths=source_relative \
            ${dir}/*.proto; \
    done

# gRPC server builder

FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.19 AS grpc-server-builder
ARG TARGETARCH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=grpc-gen /build/pkg/pb pkg/pb
RUN GOARCH=$TARGETARCH go build -o extend-event-handler

# Extend Event Handler app

FROM alpine:3.19
WORKDIR /app
COPY --from=grpc-server-builder /build/extend-event-handler .
# gRPC gateway port and Prometheus /metrics port
EXPOSE 6565 8080
CMD [ "/app/extend-event-handler" ]
