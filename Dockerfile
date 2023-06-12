FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.0.0 as proto
WORKDIR /build
COPY pkg/proto pkg/proto
RUN mkdir -p pkg/pb
RUN protoc --proto_path=pkg/proto \
    	    --go_out=pkg/pb \
    		--go_opt=paths=source_relative \
    		--go-grpc_out=pkg/pb \
    		--go-grpc_opt=paths=source_relative \
            pkg/proto/accelbyte-async-api/iam/oauth/v1/*.proto


FROM --platform=$BUILDPLATFORM golang:1.20-alpine as builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=proto /build/pkg/pb pkg/pb
RUN env GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o extend-event-listener-go_$TARGETOS-$TARGETARCH


FROM alpine:3.17.0
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
#ADD data data
COPY --from=builder /build/extend-event-listener-go_$TARGETOS-$TARGETARCH extend-event-listener-go
# Plugin arch gRPC server port
EXPOSE 6565
# Prometheus /metrics web server port
EXPOSE 8080
CMD [ "/app/extend-event-listener-go" ]
