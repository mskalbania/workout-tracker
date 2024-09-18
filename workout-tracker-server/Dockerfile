FROM golang:alpine AS builder

RUN apk add --no-cache protobuf git \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

#oprimizing build times
COPY workout-tracker-server/go.mod workout-tracker-server/go.sum /source/
WORKDIR /source
RUN go mod download

#generating proto+grpc stuff, copy to go src so it is available for the build
COPY proto /source/proto
COPY protoc.sh /source/protoc.sh
RUN chmod +x /source/protoc.sh \
    && /source/protoc.sh  \
    && cp -r /source/proto /usr/local/go/src/proto

ADD workout-tracker-server/ /source
RUN go build -o workout-tracker-server

FROM alpine:latest
COPY --from=builder /source/workout-tracker-server /app/workout-tracker-server
WORKDIR /app
CMD ["./workout-tracker-server"]