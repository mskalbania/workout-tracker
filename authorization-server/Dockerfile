FROM golang:alpine AS builder

RUN apk add --no-cache protobuf git curl \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
  && go install github.com/envoyproxy/protoc-gen-validate@latest

#oprimizing build times
COPY authorization-server/go.mod authorization-server/go.sum /source/
WORKDIR /source
RUN go mod download

#generating proto+grpc stuff, copy to go src so it is available for the build
COPY proto/auth/v1/auth.proto /source/proto/auth/v1/auth.proto
COPY protoc-auth-server.sh /source/protoc.sh
RUN chmod +x /source/protoc.sh \
    && /source/protoc.sh  \
    && cp -r /source/proto /usr/local/go/src/proto

ADD authorization-server/ /source
RUN go build -o authorization-server

FROM alpine:latest
COPY --from=builder /source/authorization-server /app/authorization-server
WORKDIR /app
CMD ["./authorization-server"]