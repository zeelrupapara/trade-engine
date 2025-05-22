news-gen:
	@protoc \
		--proto_path=protobuf "protobuf/news.proto" \
		--go_out=protobuf/genproto/news --go_opt=paths=source_relative \
  	--go-grpc_out=protobuf/genproto/news --go-grpc_opt=paths=source_relative
	