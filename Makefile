.PHONY: test
test:
	protoc --proto_path=. --proto_path=./test/third_party --go_out=paths=source_relative:. --go-http_out=paths=source_relative,frame=echo:. ./test/*.proto

.DEFAULT_GOAL := test
