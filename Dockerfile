# gRPC gen

FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.1.0 AS grpc-gen
WORKDIR /build
COPY pkg/proto pkg/proto
COPY proto.sh .
RUN mkdir -p gateway/apidocs pkg/pb
RUN bash proto.sh

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
