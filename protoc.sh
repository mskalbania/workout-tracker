#used to generate grpc code
authPath=./proto/auth/v1
workoutPath=./proto/workout/v1
protoc --go_out=$authPath --go-grpc_out=$authPath  $authPath/auth.proto
protoc --go_out=$workoutPath --go-grpc_out=$workoutPath  $workoutPath/workout.proto
