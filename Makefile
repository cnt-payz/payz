PROTO_SERVICE=?
PROTO_VERSION=v1
PROTO_CLIENT=?

proto:
	protoc --proto_path=./contracts \
		--go_out=./$(PROTO_SERVICE)-service/api --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_SERVICE)-service/api --go-grpc_opt=paths=source_relative \
		./contracts/$(PROTO_SERVICE)/$(PROTO_VERSION)/$(PROTO_SERVICE).proto

protoclient:
	protoc --proto_path=./contracts \
		--go_out=./$(PROTO_CLIENT)/api --go_opt=paths=source_relative \
		--go-grpc_out=./$(PROTO_CLIENT)/api --go-grpc_opt=paths=source_relative \
		./contracts/$(PROTO_SERVICE)/$(PROTO_VERSION)/$(PROTO_SERVICE).proto