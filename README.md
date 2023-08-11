# protoc-gen-go-http

step1
```
go get github.com/cnartlu/protoc-gen-go-http
```
step2
```
go install github.com/cnartlu/protoc-gen-go-http
```
step3
```
protoc --proto_path=. --proto_path=./third_party --go_out=paths=source_relative:. --go-http_out=paths=source_relative,frame=echo:. --go-grpc_out=paths=source_relative:. *.proto
```

---
# 参数
frame 框架
echo    已支持
gin     已支持
fiber   待支持 https://docs.gofiber.io/