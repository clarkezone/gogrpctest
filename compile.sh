#need to make sure we have a gopath: export GOPATH=$HOME/go export PATH=$PATH:$GOPATH/bin
#go get -u github.com/golang/protobuf/protoc-gen-go
protoc grpctest.proto -I jamestestrpc/ grpctest.proto --go_out=plugins=grpc:jamestestrpc
