#used to generate grpc code
protoc --go_out=./authorization-server --go-grpc_out=./authorization-server  ./proto/authorization-server/v1/authorization-server.proto
