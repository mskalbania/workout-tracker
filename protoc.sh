#used to generate grpc code
authPath=./proto/auth/v1
protoc --go_out=$authPath --go-grpc_out=$authPath  $authPath/auth.proto
