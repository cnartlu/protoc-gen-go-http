package gin

import (
	"fmt"
	"net/http"

	"github.com/cnartlu/protoc-gen-go-http/genhttp/frames"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func init() {
	frames.RegisterFrame(g{})
}

type g struct{}

func (g) Name() string {
	return "gin"
}

var (
	importContext = protogen.GoImportPath("context")
	importNetHttp = protogen.GoImportPath("net/http")
	importGin     = protogen.GoImportPath("github.com/gin-gonic/gin")
	importBinding = protogen.GoImportPath("github.com/gin-gonic/gin/binding")
	importProto   = protogen.GoImportPath("google.golang.org/protobuf/proto")
	importErrors  = protogen.GoImportPath("errors")
)

var (
	// unique func implementation
	uniqueFuncs = map[string]struct{}{}
)

func (g) Generate(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) error {
	methods := make([]frames.MethodDesc, 0, len(service.Methods))
	protogenMethods := make([]*protogen.Method, 0, len(service.Methods))
	{
		methodMap := map[string]struct{}{}
		for idx := range service.Methods {
			m := service.Methods[idx]
			if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
				continue
			}
			rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
			if ok {
				methods = append(methods, frames.NewMethodDesc(m, rule))
				if rule != nil {
					for idx, rule := range rule.AdditionalBindings {
						methods = append(methods, frames.NewMethodDesc(m, rule).AddNum(idx+1))
					}
				}
			} else if !omitempty {
				methods = append(methods, frames.NewMethodDesc(m, nil))
			}
			if _, ok := methodMap[string(m.Desc.FullName())]; !ok {
				protogenMethods = append(protogenMethods, m)
			}
			methodMap[string(m.Desc.FullName())] = struct{}{}
		}
	}
	if len(protogenMethods) < 1 {
		return nil
	}

	createUniquefunc(gen, file, g)
	generateUnimplemented(g, service, protogenMethods)
	generateRouterMethods(g, service, methods)
	generateHandlerMethods(g, service, methods)
	return nil
}

// createUniquefunc package unique func
func createUniquefunc(plugin *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if _, ok := uniqueFuncs[string(file.GoImportPath)]; ok {
		return
	}
	uniqueFuncs[string(file.GoImportPath)] = struct{}{}

	bindingValidator := g.QualifiedGoIdent(importBinding.Ident("Validator"))
	ginContext := g.QualifiedGoIdent(importGin.Ident("Context"))

	g.P(`
	var (
		BindGinTagName = "json"
		// GinResponseBodyKey represents the response content key
		GinResponseBodyKey = "_gin-gonic/gin/responsebodykey"
		// GinBindRequestBody binds the body parameter
		GinBindRequestBody  = _ginBindRequestBody
	)

	func SetGinBindRequestBody(f func(*`, ginContext, `, any) error) {
		GinBindRequestBody = f
	}

	// _ginBindRequestBody default bind handler
	func _ginBindRequestBody(c *`, ginContext, `, req any) error {
		return c.Bind(req)
	}

	func ginValidate(obj any) error{
		if `, bindingValidator, ` == nil{
			return nil
		}
		return `, bindingValidator, `.ValidateStruct(obj)
	}
`)

	generateBindParamsFunc(g)
	generateBindQueryFunc(g)
}

func generateBindParamsFunc(g *protogen.GeneratedFile) {
	ginContext := g.QualifiedGoIdent(importGin.Ident("Context"))
	ginErrorTypeBind := g.QualifiedGoIdent(importGin.Ident("ErrorTypeBind"))
	protoMessage := g.QualifiedGoIdent(importProto.Ident("Message"))
	httpStatusBadRequest := g.QualifiedGoIdent(importNetHttp.Ident("StatusBadRequest"))
	bindingMapFormWithTag := g.QualifiedGoIdent(importBinding.Ident("MapFormWithTag"))

	g.P(`func _Bind_Gin_Params(c *`, ginContext, `, req `, protoMessage, `) error {
		m := make(map[string][]string)
		for _, v := range c.Params {
			m[v.Key] = []string{v.Value}
		}
		return `, bindingMapFormWithTag, `(req, m, BindGinTagName)
	}

	func _Abort_Bind_Gin_Params(c *`, ginContext, `, req `, protoMessage, `) error {
		if err := _Bind_Gin_Params(c, req); err != nil {
			c.Status(`, httpStatusBadRequest, `)
			c.Abort()
			return c.Error(err).SetType(`, ginErrorTypeBind, `) //nolint: errcheck
		}
		return nil
	}`)
	g.P()
}

func generateBindQueryFunc(g *protogen.GeneratedFile) {
	ginContext := g.QualifiedGoIdent(importGin.Ident("Context"))
	ginErrorTypeBind := g.QualifiedGoIdent(importGin.Ident("ErrorTypeBind"))
	protoMessage := g.QualifiedGoIdent(importProto.Ident("Message"))
	httpStatusBadRequest := g.QualifiedGoIdent(importNetHttp.Ident("StatusBadRequest"))
	bindingMapFormWithTag := g.QualifiedGoIdent(importBinding.Ident("MapFormWithTag"))

	g.P(`func _Bind_Gin_Query(c *`, ginContext, `, req `, protoMessage, `) error {
		query := c.Request.URL.Query()
		for _, v := range c.Params {
			if query.Get(v.Key) == "" {
				query.Set(v.Key, v.Value)
			}
		}
		return `, bindingMapFormWithTag, `(req, query, BindGinTagName)
	}

	func _Abort_Bind_Gin_Query(c *`, ginContext, `, req `, protoMessage, `) error {
		if err := _Bind_Gin_Query(c, req); err != nil {
			c.Status(`, httpStatusBadRequest, `)
			c.Abort()
			return c.Error(err).SetType(`, ginErrorTypeBind, `) //nolint: errcheck
		}
		return nil
	}`)
}

// generateUnimplemented
func generateUnimplemented(g *protogen.GeneratedFile, service *protogen.Service, methods []*protogen.Method) {
	contextIdent := g.QualifiedGoIdent(importContext.Ident("Context"))
	ginIdent := g.QualifiedGoIdent(importGin.Ident(""))
	errorsIdent := g.QualifiedGoIdent(importErrors.Ident(""))
	httpIdent := g.QualifiedGoIdent(importNetHttp.Ident(""))

	g.P("// ", service.GoName, "GinServer is the server API for ", service.GoName, " service.")
	g.P("// All implementations must embed Unimplemented", service.GoName, "GinServer")
	g.P("// for forward compatibility")
	g.P("type ", service.GoName, "GinServer interface {")
	for _, m := range methods {
		g.P(m.GoName, "(ctx ", contextIdent, ", req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error)")
	}
	g.P("mustEmbedUnimplemented", service.GoName, "Server()")
	g.P("}")

	// unimplemented server
	g.P("// Unimplemented", service.GoName, "GinServer must be embedded to have forward compatible implementations.")
	g.P("type Unimplemented", service.GoName, "GinServer struct {}")
	for _, m := range methods {
		g.P("func (Unimplemented", service.GoName, "GinServer) ", m.GoName, "(ctx ", contextIdent, ", req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error) {")
		g.P("return nil, ", ginIdent, "Error{Type: ", ginIdent, "ErrorTypePublic, Err: ", errorsIdent, "New(", httpIdent, "StatusText(", httpIdent, "StatusNotImplemented))}")
		g.P("}")
	}
	g.P("func (Unimplemented", service.GoName, "GinServer) mustEmbedUnimplemented", service.GoName, "Server() {}")

	// unsafe server
	g.P("// Unsafe", service.GoName, "GinServer may be embedded to opt out of forward compatibility for this service.")
	g.P("// Use of this interface is not recommended, as added methods to ", service.GoName, "GinServer will")
	g.P("// result in compilation errors.")
	g.P("type Unsafe", service.GoName, "GinServer interface {")
	g.P("mustEmbedUnimplemented", service.GoName, "Server()")
	g.P("}")
	g.P()
}

// generateRouterMethods 注册路由
func generateRouterMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []frames.MethodDesc) {
	ginIRoutesIdent := g.QualifiedGoIdent(importGin.Ident("IRoutes"))
	g.P("type ", service.GoName, "GinRouter = ", ginIRoutesIdent)
	g.P()
	g.P("func Register", service.GoName, "GinServer(r ", service.GoName, "GinRouter, srv ", service.GoName, "GinServer) {")
	for _, m := range methods {
		g.P(fmt.Sprintf(`r.%s("%s", _%s_%s%d_Gin_Handler(srv))`, m.MethodName, m.Path, service.GoName, m.GoName, m.Num))
	}
	g.P("}")
	g.P()
}

// generateHandlerMethods 生成执行方法
func generateHandlerMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []frames.MethodDesc) {
	for _, m := range methods {
		generateHandlerMethod(g, service, m)
	}
}

func generateHandlerMethod(g *protogen.GeneratedFile, service *protogen.Service, m frames.MethodDesc) {
	g.P(fmt.Sprintf("func _%s_%s%d_Gin_Handler(srv %sGinServer) gin.HandlerFunc {", service.GoName, m.GoName, m.Num, service.GoName))
	generateHandlerMethodHandler(g, service, m)
	g.P("}")
	g.P()
}

func generateHandlerMethodHandler(g *protogen.GeneratedFile, service *protogen.Service, m frames.MethodDesc) {
	ginContextIdent := g.QualifiedGoIdent(importGin.Ident("Context"))
	statusInternalServerErrorIdent := g.QualifiedGoIdent(importNetHttp.Ident("StatusInternalServerError"))
	errorTypeAnyIdent := g.QualifiedGoIdent(importGin.Ident("ErrorTypeAny"))

	g.P(`return func(c *`, ginContextIdent, `) {
		req := new(`, m.Input.GoIdent, `)`)
	generateBindRequest(g, m)
	g.P(`res, err := srv.`, m.GoName, `(c, req)
	if err != nil {
		c.Status(`, statusInternalServerErrorIdent, `)
		_ = c.Error(err).SetType(`, errorTypeAnyIdent, `) //nolint: errcheck
		return
	}
	c.Set(GinResponseBodyKey, res`, m.ResponseBody, `)`)
	g.P("}")
}

func generateBindRequest(g *protogen.GeneratedFile, m frames.MethodDesc) {
	switch m.Input.Desc.FullName() {
	case "google.protobuf.Empty":
		return
	default:
	}
	ginErrorTypeBind := g.QualifiedGoIdent(importGin.Ident("ErrorTypeBind"))
	httpStatusBadRequest := g.QualifiedGoIdent(importNetHttp.Ident("StatusBadRequest"))
	// get and delete bind query and path params
	if len(m.Params) > 0 {
		g.P(`if err := _Abort_Bind_Gin_Params(c, req); err != nil {
			return
		}`)
	}

	switch m.MethodName {
	case http.MethodGet, http.MethodDelete:
		g.P(`if err := _Abort_Bind_Gin_Query(c, req`, m.Body, `); err != nil {
			return
		}`)
	default:
		g.P(`if err := GinBindRequestBody(c, req); err != nil {
			c.Status(`, httpStatusBadRequest, `)
			_ = c.Error(err).SetType(`, ginErrorTypeBind, `) //nolint: errcheck
			return
	}`)
	}

	g.P(`if err := ginValidate(req); err != nil {
		c.Status(`, httpStatusBadRequest, `)
		_ = c.Error(err).SetType(`, ginErrorTypeBind, `) //nolint: errcheck
		return
}`)
}
