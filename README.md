# gRPC File Transfer

This project shows how to send files from a client to the server.\
For demo purpose, we allow YAML/JSON and certificate files only.

## Prerequisites

- [orotoc](https://github.com/protocolbuffers/protobuf/releases)
- [protoc-gen-go](https://github.com/golang/protobuf)
- [Common Protos](https://github.com/googleapis/api-common-protos)

## Protobuf

Download standdard protofiles:
```
$ go get -v https://github.com/googleapis/api-common-protos
```

Compile protobuf:
```
$ cd proto
$ protoc --go_out=plugins=grpc:. api.proto -I="$GOPATH/src/github.com/googleapis/api-common-protos" -I="./"
```

Compile server and client:
```
$ cd ../
$ go build -o server cmd/srv/main.go
$ go build -o client cmd/cli/main.go
```

Start server:
```
$ ./server
```

Send file:
```
$ echo '{"hello":"world"}' > my_file.json
$ ./client --api.address="127.0.0.1:8081" file push my_file.json my_namespace:my_file.json
```

## ToDo

- Include API address in command argument.
- Add TLS feature.
- Implement `cat` command.
- Write file instead of printing it to stdout.
