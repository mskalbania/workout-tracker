FROM golang:alpine AS builder

RUN apk add --no-cache protobuf git curl \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
  && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

#oprimizing build times
COPY grpc-gateway/go.mod grpc-gateway/go.sum /source/
WORKDIR /source
RUN go mod download

#generating proto+grpc stuff, copy to go src so it is available for the build
COPY proto /source/proto
COPY protoc-all.sh /source/protoc.sh
RUN chmod +x /source/protoc.sh \
    && /source/protoc.sh  \
    && cp -r /source/proto /usr/local/go/src/proto

ADD grpc-gateway/ /source
RUN go build -o grpc-gateway

FROM alpine:latest
COPY --from=builder /source/grpc-gateway /app/grpc-gateway
WORKDIR /app
CMD ["./grpc-gateway"]