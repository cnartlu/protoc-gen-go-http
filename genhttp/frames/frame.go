package frames

import (
	"google.golang.org/protobuf/compiler/protogen"
)

var registeredFrames = make(map[string]Frame)

type Frame interface {
	Name() string
	Generate(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) error
}

// RegisterFrame registers the provided Frame for use with all Transport clients and servers.
func RegisterFrame(codec Frame) {
	if codec == nil {
		panic("cannot register a nil Frame")
	}
	if codec.Name() == "" {
		panic("cannot register Frame with empty string result for Name()")
	}
	registeredFrames[codec.Name()] = codec
}

// GetFrame gets a registered Frame by content-subtype, or nil if no Frame is
// registered for the content-subtype.
//
// The content-subtype is expected to be lowercase.
func GetFrame(name string) Frame {
	return registeredFrames[name]
}
