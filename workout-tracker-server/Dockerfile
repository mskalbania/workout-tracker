FROM golang:alpine AS builder

RUN apk add --no-cache protobuf git curl \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
  && go install github.com/envoyproxy/protoc-gen-validate@latest

#oprimizing build times
COPY workout-tracker-server/go.mod workout-tracker-server/go.sum /source/
WORKDIR /source
RUN go mod download

#generating proto+grpc stuff, copy to go src so it is available for the build
COPY proto/workout/v1/workout.proto /source/proto/workout/v1/workout.proto
COPY protoc-workout-tracker-server.sh /source/protoc.sh
RUN chmod +x /source/protoc.sh \
    && /source/protoc.sh  \
    && cp -r /source/proto /usr/local/go/src/proto

ADD workout-tracker-server/ /source
RUN go build -o workout-tracker-server

FROM alpine:latest
COPY --from=builder /source/workout-tracker-server /app/workout-tracker-server
WORKDIR /app
CMD ["./workout-tracker-server"]