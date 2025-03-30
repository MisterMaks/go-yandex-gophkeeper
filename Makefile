.PHONY: protoc
protoc:
	@echo "-- compiling .proto files"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
	--go-grpc_opt=paths=source_relative api/proto/service/service.proto;
