Run these commands from `pariksha/backend` directory

## Compile all proto files

> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/*.proto"

## Compile specific proto files

<!-- Common -->
> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/common.proto"

<!-- Auth -->
> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/auth.proto"

<!-- User -->
> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/user.proto"

<!-- Paper -->
> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/paper.proto"

<!-- Exam -->
> protoc --proto_path=common/proto --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/exam.proto"
