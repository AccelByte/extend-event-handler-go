FROM alpine:3.18
WORKDIR /workspace
RUN apk --no-cache add bash curl docker-cli-buildx git go jq make \
        && apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing grpcurl
COPY --from=ngrok/ngrok:latest /bin/ngrok /bin/ngrok
