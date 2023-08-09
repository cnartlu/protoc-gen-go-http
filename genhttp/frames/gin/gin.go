package gin

import "google.golang.org/protobuf/compiler/protogen"

type Gin struct {
}

func (e *Gin) ConvertPathToVars(input *protogen.Message, p string) map[string]string {
	return nil
}

func (g *Gin) GenerateFile() {

}
