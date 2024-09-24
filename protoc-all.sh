mkdir -p ./proto/google/api
mkdir -p ./proto/google/protobuf
curl --request GET \
     --url 'https://raw.githubusercontent.com/googleapis/googleapis/refs/heads/master/google/api/http.proto' \
     --output './proto/google/api/http.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/googleapis/googleapis/refs/heads/master/google/api/annotations.proto' \
     --output './proto/google/api/annotations.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/googleapis/googleapis/refs/heads/master/google/api/field_behavior.proto' \
     --output './proto/google/api/field_behavior.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/googleapis/googleapis/refs/heads/master/google/api/httpbody.proto' \
     --output './proto/google/api/httpbody.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/protocolbuffers/protobuf/refs/heads/main/src/google/protobuf/descriptor.proto' \
     --output './proto/google/protobuf/descriptor.proto'

authPath=./proto/auth/v1
workoutPath=./proto/workout/v1
protoc -I ./proto -I ./proto/google/api --go_out=$authPath --go-grpc_out=$authPath --grpc-gateway_out=$authPath $authPath/auth.proto
protoc -I ./proto -I ./proto/google/api --go_out=$workoutPath --go-grpc_out=$workoutPath --grpc-gateway_out=$workoutPath $workoutPath/workout.proto
