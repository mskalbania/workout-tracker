mkdir -p ./proto/google/api
mkdir -p ./proto/google/protobuf
mkdir -p ./proto/validate
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
curl --request GET \
     --url 'https://raw.githubusercontent.com/protocolbuffers/protobuf/refs/heads/main/src/google/protobuf/empty.proto' \
     --output './proto/google/protobuf/empty.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/protocolbuffers/protobuf/refs/heads/main/src/google/protobuf/field_mask.proto' \
     --output './proto/google/protobuf/field_mask.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/protocolbuffers/protobuf/refs/heads/main/src/google/protobuf/duration.proto' \
     --output './proto/google/protobuf/duration.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/protocolbuffers/protobuf/refs/heads/main/src/google/protobuf/timestamp.proto' \
     --output './proto/google/protobuf/timestamp.proto'
curl --request GET \
     --url 'https://raw.githubusercontent.com/bufbuild/protoc-gen-validate/refs/heads/main/validate/validate.proto' \
     --output './proto/validate/validate.proto'

workoutPath=./proto/workout/v1

protoc -I ./proto -I ./proto/google/api \
--go_out=$workoutPath \
--go-grpc_out=$workoutPath \
--validate_out="lang=go:$workoutPath" \
$workoutPath/workout.proto
