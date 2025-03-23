Run these commands from the root, i.e. pariksha directory

<!-- Auth -->
> protoc --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/auth.proto"

<!-- Paper -->
> protoc --go_out="./common/pkg" --go-grpc_out="./common/pkg" "common/proto/paper.proto"
