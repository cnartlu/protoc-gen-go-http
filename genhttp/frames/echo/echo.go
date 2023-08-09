package echo

import (
	"fmt"

	"github.com/cnartlu/protoc-gen-go-http/genhttp/frames"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func init() {
	frames.RegisterFrame(e{})
}

type e struct{}

func (e) Name() string {
	return "echo"
}

func (e) Generate(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) error {
	methods := make([]methodDesc, 0, len(service.Methods))
	methodMap := map[string]*protogen.Method{}
	for idx := range service.Methods {
		m := service.Methods[idx]
		if m.Desc.IsStreamingClient() || m.Desc.IsStreamingServer() {
			continue
		}
		rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		if ok {
			methods = append(methods, NewMethodDesc(m, rule))
			if rule != nil {
				for idx, rule := range rule.AdditionalBindings {
					methods = append(methods, NewMethodDesc(m, rule).AddNum(idx+1))
				}
			}
		} else if !omitempty {
			methods = append(methods, NewMethodDesc(m, nil))
		}
		methodMap[string(m.Desc.FullName())] = m
	}
	if len(methodMap) < 1 {
		return nil
	}

	g.QualifiedGoIdent(protogen.GoImportPath("net/http").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("context").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("strings").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("github.com/labstack/echo/v4").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("google.golang.org/protobuf/proto").Ident(""))

	// 定义Http服务接口
	g.P("// ", service.GoName, "HttpServer is the server API for ", service.GoName, " service.")
	g.P("// All implementations must embed Unimplemented", service.GoName, "HttpServer")
	g.P("// for forward compatibility")
	g.P("type ", service.GoName, "HttpServer interface {")
	for _, m := range methodMap {
		g.P(m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error)")
	}
	g.P("mustEmbedUnimplemented", service.GoName, "HttpServer()")
	g.P("}")

	// Unimplemented 定义
	g.P("// Unimplemented", service.GoName, "HttpServer must be embedded to have forward compatible implementations.")
	{
		g.P("type Unimplemented", service.GoName, "HttpServer struct {")
	}
	g.P("}")
	for _, m := range methodMap {
		g.P("func (Unimplemented", service.GoName, "HttpServer) ", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error) {")
		{
			g.P("return nil, v4.ErrNotImplemented")
		}
		g.P("}")
	}
	g.P("func (Unimplemented", service.GoName, "HttpServer) mustEmbedUnimplemented", service.GoName, "HttpServer() {}")

	// 注册
	g.P("// Unsafe", service.GoName, "HttpServer may be embedded to opt out of forward compatibility for this service.")
	g.P("// Use of this interface is not recommended, as added methods to ", service.GoName, "HttpServer will")
	g.P("// result in compilation errors.")
	g.P("type Unsafe", service.GoName, "HttpServer interface {")
	g.P("mustEmbedUnimplemented", service.GoName, "HttpServer()")
	g.P("}")
	g.P()

	generateRouterMethods(g, service, methods)
	generateHandlerMethods(g, service, methods)
	return nil
}

// generateRouterMethods 注册路由
func generateRouterMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []methodDesc) {
	g.P("type ", service.GoName, "HttpRouter interface {")
	{
		g.P("Add(method, path string, handler v4.HandlerFunc, middleware ...v4.MiddlewareFunc) *v4.Route")
	}
	g.P("}")
	g.P()
	g.P("func Register", service.GoName, "HttpServer(r ", service.GoName, "HttpRouter, srv ", service.GoName, "HttpServer) {")
	for _, m := range methods {
		g.P(fmt.Sprintf(`r.Add("%s", "%s", _%s_%s%d_HTTP_Handler(srv))`, m.MethodName, m.Path, service.GoName, m.GoName, m.Num))
	}
	g.P("}")
	g.P()
}

// generateHandlerMethods 生成执行方法
func generateHandlerMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []methodDesc) {
	for _, m := range methods {
		generateHandlerMethod(g, service, m)
	}
}

func generateHandlerMethod(g *protogen.GeneratedFile, service *protogen.Service, m methodDesc) {
	g.P(fmt.Sprintf("func _%s_%s%d_HTTP_Handler(srv %sHttpServer) v4.HandlerFunc {", service.GoName, m.GoName, m.Num, service.GoName))
	// start handler
	{
		g.P("return func(c v4.Context) error {")
		g.P("req := new(", m.Input.GoIdent, ")")
		// 判断输入是否为空
		{
			switch m.Input.Desc.FullName() {
			case "google.protobuf.Empty":
				break
			default:
				g.P("if err := c.Bind(req" + m.Body + "); err != nil { return err }")
			}
		}
		// 参数验证
		{
			g.P("// param validate")
			g.P("if err:= c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered { return err }")
		}
		// 请求响应结果
		g.P("// response body")
		g.P("res, err := srv.", m.GoName, "(c.Request().Context(), req)")
		g.P("if err != nil { return err }")
		{
			resSuffix := ""
			if m.ResponseBody != "" {
				resSuffix = "p"
				g.P("// return res", m.ResponseBody, " value")
				g.P("res", resSuffix, " := res", m.ResponseBody)
			}
			g.P("accept :=strings.ToLower(c.Request().Header.Get(\"Accept\"))")
			g.P("switch {")
			{
				g.P(`case strings.Contains(accept, "application/x-protobuf") || strings.Contains(accept, "application/protobuf"):`)
				g.P("bs,_ := proto.Marshal(res", resSuffix, ")")
				g.P(`return c.Blob(http.StatusOK, "application/x-protobuf", bs)`)
				g.P(`case accept == "*/*" || strings.Contains(accept, "application/json"):`)
				g.P("return c.JSON(http.StatusOK, res", resSuffix, ")")
				g.P(`case strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml"):`)
				g.P("return c.XML(http.StatusOK, res", resSuffix, ")")
				g.P(`default:`)
				g.P("return c.JSON(http.StatusOK, res", resSuffix, ")")
			}
			g.P("}")
		}
		g.P("}")
	}
	// end
	g.P("}")
	g.P()
}
