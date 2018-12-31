sudo apt-get install unzip
wget -o foo3.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
unzip protoc-3.6.1-linux-x86_64.zip -d protoc
rm foo3.zip
protoc/bin/protoc grpctest.proto -I jamestestrpc/ grpctest.proto --go_out=plugins=grpc:jamestestrpc

