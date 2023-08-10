package main

import (
	"flag"

	"github.com/cnartlu/protoc-gen-go-http/genhttp"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	_ "github.com/cnartlu/protoc-gen-go-http/genhttp/frames/echo"
	_ "github.com/cnartlu/protoc-gen-go-http/genhttp/frames/gin"
)

var (
	version   = "v0.1.1"
	frame     = flag.String("frame", "echo", "the http framework used")
	omitempty = flag.Bool("omitempty", true, "omit if google.api is empty")
)

func main() {
	flag.Parse()
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			return genhttp.New().
				SetFile(f).
				SetFrame(*frame).
				SetOmitempty(*omitempty).
				SetPlugin(gen).
				SetVersion(version).
				Generate()
		}
		return nil
	})
}
