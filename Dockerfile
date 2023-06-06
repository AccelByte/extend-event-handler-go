FROM --platform=$BUILDPLATFORM rvolosatovs/protoc:4.0.0 as proto
WORKDIR /build
COPY pkg/proto pkg/proto
RUN mkdir -p pkg/pb
RUN protoc --proto_path=pkg/proto \
    	    --go_out=pkg/pb \
    		--go_opt=paths=source_relative \
    		--go-grpc_out=pkg/pb \
    		--go-grpc_opt=paths=source_relative \
    		--grpc-gateway_out=pkg/pb \
    		--grpc-gateway_opt=logtostderr=true \
    		--grpc-gateway_opt=paths=source_relative \
    		--openapiv2_out=pkg/pb \
    		--openapiv2_opt=logtostderr=true \
    		pkg/proto/*.proto


FROM --platform=$BUILDPLATFORM golang:1.20-alpine as builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=proto /build/pkg/pb pkg/pb
RUN env GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o extend-grpc-event-handler-go_$TARGETOS-$TARGETARCH


FROM alpine:3.17.0
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
#ADD data data
COPY --from=builder /build/extend-grpc-event-handler-go_$TARGETOS-$TARGETARCH extend-grpc-event-handler-go
# Plugin arch gRPC server port
EXPOSE 6565
# Prometheus /metrics web server port
EXPOSE 8080
CMD [ "/app/extend-grpc-event-handler-go" ]