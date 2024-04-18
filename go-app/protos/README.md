protoc --go_out=. --go_opt=paths=source_relative \
 --go-grpc_out=. --go-grpc_opt=paths=source_relative \
 simple.proto

grpcurl -plaintext localhost:7070 list
grpcurl -plaintext localhost:7070 list protos.DeviceConfigurator
grpcurl -plaintext localhost:7070 describe protos.DeviceConfigurator

grpcurl -plaintext -format json -d '{"configWiFi": "ciao"}' \
 localhost:7070 protos.DeviceConfigurator.UpdateDeviceConfig
