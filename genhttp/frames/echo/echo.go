package echo

import (
	"fmt"
	"strconv"

	"github.com/cnartlu/protoc-gen-go-http/genhttp/frames"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func init() {
	frames.RegisterFrame(e{})
}

var (
	importEcho    = protogen.GoImportPath("github.com/labstack/echo/v4")
	importProto   = protogen.GoImportPath("google.golang.org/protobuf/proto")
	importStrings = protogen.GoImportPath("strings")
	importContext = protogen.GoImportPath("context")
	importHttp    = protogen.GoImportPath("net/http")
)

var (
	// unique func implementation
	uniqueFuncs = map[string]struct{}{}
)

type e struct{}

func (e) Name() string {
	return "echo"
}

func (e) Generate(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) error {
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

	g.QualifiedGoIdent(importStrings.Ident(""))
	// write parseAccept func
	g.P(`func echoParseAccept(acceptHeader string) []string {
	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if i := strings.IndexByte(part, ';'); i > 0 {
			part = part[:i]
		}
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}`)
	g.P()

	contextIdent := g.QualifiedGoIdent(importEcho.Ident("Context"))
	_ = g.QualifiedGoIdent(importProto.Ident("Message"))
	g.QualifiedGoIdent(importHttp.Ident(""))
	g.P("func _OutEchoResponseHandler(c ", contextIdent, ", res any) error {")
	g.P("accepted := echoParseAccept(c.Request().Header.Get(", strconv.Quote("Accept"), "))")
	g.P(`for _, accept := range accepted {
		switch accept {
		case "application/json":
			return c.JSON(http.StatusOK, res)
		case "application/xml","text/xml":
			return c.XML(http.StatusOK, res)
		case "application/x-protobuf", "application/protobuf":
			bs, _ := proto.Marshal(res.(proto.Message))
			return c.Blob(http.StatusOK, "application/x-protobuf", bs)
		default:
		}
	}
	return c.JSON(http.StatusOK, res)`)
	g.P("}")

}

func generateUnimplemented(g *protogen.GeneratedFile, service *protogen.Service, methods []*protogen.Method) {
	contextIdent := g.QualifiedGoIdent(importContext.Ident("Context"))
	ErrNotImplementedIdent := g.QualifiedGoIdent(importEcho.Ident("ErrNotImplemented"))

	// Unimplemented server interface
	g.P("// ", service.GoName, "EchoServer is the server API for ", service.GoName, " service.")
	g.P("// All implementations must embed Unimplemented", service.GoName, "EchoServer")
	g.P("// for forward compatibility")
	g.P("type ", service.GoName, "EchoServer interface {")
	for _, m := range methods {
		g.P(m.GoName, "(ctx ", contextIdent, ", req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error)")
	}
	g.P("mustEmbedUnimplemented", service.GoName, "Server()")
	g.P("}")

	// Unimplemented server
	g.P("// Unimplemented", service.GoName, "EchoServer must be embedded to have forward compatible implementations.")
	g.P("type Unimplemented", service.GoName, "EchoServer struct {")
	g.P("}")

	for _, m := range methods {
		g.P("func (Unimplemented", service.GoName, "EchoServer) ", m.GoName, "(ctx ", contextIdent, ", req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error) {")
		g.P("return nil, ", ErrNotImplementedIdent)
		g.P("}")
	}
	g.P("func (Unimplemented", service.GoName, "EchoServer) mustEmbedUnimplemented", service.GoName, "Server() {}")

	// unsafe server
	g.P("// Unsafe", service.GoName, "EchoServer may be embedded to opt out of forward compatibility for this service.")
	g.P("// Use of this interface is not recommended, as added methods to ", service.GoName, "EchoServer will")
	g.P("// result in compilation errors.")
	g.P("type Unsafe", service.GoName, "EchoServer interface {")
	g.P("mustEmbedUnimplemented", service.GoName, "Server()")
	g.P("}")
	g.P()
}

// generateRouterMethods 注册路由
func generateRouterMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []frames.MethodDesc) {
	routeIdent := g.QualifiedGoIdent(importEcho.Ident("Route"))
	handlerFuncIdent := g.QualifiedGoIdent(importEcho.Ident("HandlerFunc"))
	middlewareFuncIdent := g.QualifiedGoIdent(importEcho.Ident("MiddlewareFunc"))

	g.P("type ", service.GoName, "HttpRouter interface {")
	g.P("Add(method, path string, handler ", handlerFuncIdent, ", middleware ...", middlewareFuncIdent, ") *", routeIdent)
	g.P("}")
	g.P()
	g.P("func Register", service.GoName, "EchoServer(r ", service.GoName, "HttpRouter, srv ", service.GoName, "EchoServer) {")
	for _, m := range methods {
		g.P(fmt.Sprintf(`r.Add("%s", "%s", _%s_%s%d_Echo_Handler(srv))`, m.MethodName, m.Path, service.GoName, m.GoName, m.Num))
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
	handlerFuncIdent := g.QualifiedGoIdent(importEcho.Ident("HandlerFunc"))
	g.P("func _", service.GoName, "_", m.GoName, m.Num, "_Echo_Handler(srv ", service.GoName, "EchoServer) ", handlerFuncIdent, " {")
	generateHandlerMethodHandler(g, service, m)
	g.P("}")
	g.P()
}

func generateHandlerMethodHandler(g *protogen.GeneratedFile, service *protogen.Service, m frames.MethodDesc) {
	contextIdent := g.QualifiedGoIdent(importEcho.Ident("Context"))
	ErrValidatorNotRegisteredIdent := g.QualifiedGoIdent(importEcho.Ident("ErrValidatorNotRegistered"))

	g.P("return func(c ", contextIdent, ") error {")
	g.P("req := new(", m.Input.GoIdent, ")")

	switch m.Input.Desc.FullName() {
	case "google.protobuf.Empty":
		break
	default:
		if len(m.Params) > 0 && m.Body != "" {
			g.P("_ = c.Bind(req)")
		}
		g.P("if err := c.Bind(req" + m.Body + "); err != nil { return err }")
	}

	g.P("if err:= c.Validate(req); err != nil && err != ", ErrValidatorNotRegisteredIdent, " { return err }")

	g.P("res, err := srv.", m.GoName, "(c.Request().Context(), req)")
	g.P("if err != nil { return err }")

	if m.ResponseBody != "" {
		g.P("return _OutEchoResponseHandler(c, res", m.ResponseBody, ")")
	} else {
		g.P("return _OutEchoResponseHandler(c, res)")
	}
	g.P("}")
}
