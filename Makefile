te-gen:
	@protoc \
		--proto_path=proto "proto/engine.proto" \
		--go_out=proto/genproto/engine --go_opt=paths=source_relative \
  	--go-grpc_out=proto/genproto/engine --go-grpc_opt=paths=source_relative
	