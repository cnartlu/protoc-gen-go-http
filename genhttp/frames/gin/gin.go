package gin

import (
	"fmt"

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

func (g) Generate(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, omitempty bool) error {
	methods := make([]frames.MethodDesc, 0, len(service.Methods))
	methodMap := map[string]*protogen.Method{}
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
		methodMap[string(m.Desc.FullName())] = m
	}
	if len(methodMap) < 1 {
		return nil
	}

	g.QualifiedGoIdent(protogen.GoImportPath("net/http").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("context").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("strings").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("errors").Ident(""))
	g.QualifiedGoIdent(protogen.GoImportPath("github.com/gin-gonic/gin").Ident(""))

	// 定义Http服务接口
	g.P("// ", service.GoName, "GinServer is the server API for ", service.GoName, " service.")
	g.P("// All implementations must embed Unimplemented", service.GoName, "GinServer")
	g.P("// for forward compatibility")
	g.P("type ", service.GoName, "GinServer interface {")
	for _, m := range methodMap {
		g.P(m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error)")
	}
	g.P("mustEmbedUnimplemented", service.GoName, "GinServer()")
	g.P("}")

	// Unimplemented 定义
	g.P("// Unimplemented", service.GoName, "GinServer must be embedded to have forward compatible implementations.")
	{
		g.P("type Unimplemented", service.GoName, "GinServer struct {")
	}
	g.P("}")
	for _, m := range methodMap {
		g.P("func (Unimplemented", service.GoName, "GinServer) ", m.GoName, "(ctx context.Context, req *", m.Input.GoIdent, ") (*", m.Output.GoIdent, ", error) {")
		{
			g.P("return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}")
		}
		g.P("}")
	}
	g.P("func (Unimplemented", service.GoName, "GinServer) mustEmbedUnimplemented", service.GoName, "GinServer() {}")

	// 注册
	g.P("// Unsafe", service.GoName, "GinServer may be embedded to opt out of forward compatibility for this service.")
	g.P("// Use of this interface is not recommended, as added methods to ", service.GoName, "GinServer will")
	g.P("// result in compilation errors.")
	g.P("type Unsafe", service.GoName, "GinServer interface {")
	g.P("mustEmbedUnimplemented", service.GoName, "GinServer()")
	g.P("}")
	g.P()

	generateRouterMethods(g, service, methods)
	generateHandlerMethods(g, service, methods)

	// output response
	g.P("func _Output_Gin_", service.GoName, "(c *gin.Context, res any) {")
	{
		g.P("accept :=strings.ToLower(c.GetHeader(\"Accept\"))")
		g.P("switch {")
		{
			g.P(`case strings.Contains(accept, "application/x-protobuf") || strings.Contains(accept, "application/protobuf"):`)
			g.P("c.ProtoBuf(http.StatusOK, res)")
			g.P(`case accept == "*/*" || strings.Contains(accept, "application/json"):`)
			g.P("c.JSON(http.StatusOK, res)")
			g.P(`case strings.Contains(accept, "application/xml") || strings.Contains(accept, "text/xml"):`)
			g.P("c.XML(http.StatusOK, res)")
			g.P(`default:`)
			g.P("c.JSON(http.StatusOK, res)")
		}
		g.P("}")
	}
	g.P("}")

	return nil
}

// generateRouterMethods 注册路由
func generateRouterMethods(g *protogen.GeneratedFile, service *protogen.Service, methods []frames.MethodDesc) {
	g.P("type ", service.GoName, "GinRouter = gin.IRoutes")
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
	// start handler
	{
		g.P("return func(c *gin.Context) {")
		g.P("req := new(", m.Input.GoIdent, ")")
		// 判断输入是否为空
		{
			switch m.Input.Desc.FullName() {
			case "google.protobuf.Empty":
				break
			default:
				g.P("// param bind and validate")
				g.P("if err := c.Bind(req" + m.Body + "); err != nil { return }")
			}
		}
		// 请求响应结果
		g.P("// response body")
		g.P("res, err := srv.", m.GoName, "(c, req)")
		g.P("if err != nil { c.Abort() \n c.Error(err) \n return }")
		{
			resSuffix := ""
			if m.ResponseBody != "" {
				resSuffix = "p"
				g.P("// return res", m.ResponseBody, " value")
				g.P("res", resSuffix, " := res", m.ResponseBody)
			}
			g.P("_Output_Gin_", service.GoName, "(c, res", resSuffix, ")")
		}
		g.P("}")
	}
	// end
	g.P("}")
	g.P()
}
