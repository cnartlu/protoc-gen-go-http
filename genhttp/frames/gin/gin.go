package gin

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

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
	bindingIdent     = protogen.GoImportPath("github.com/gin-gonic/gin/binding").Ident("")
	bindingJsonIdent = protogen.GoImportPath("google.golang.org/protobuf/encoding/protojson").Ident("")
	protoIdent       = protogen.GoImportPath("google.golang.org/protobuf/proto").Ident("")
)

var (
	uniqueFileNames = map[string]string{}
	// unique file implementation
	uniqueFiles = map[string]struct{}{}
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

	// Create .s assembly file, linkname skips build check

	fullname := string(file.Desc.FullName())
	uniqueFileNames[fullname] = hex.EncodeToString(md5.New().Sum([]byte(fullname)))

	createUniqueFile(gen, file)

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
	for _, m := range protogenMethods {
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
	for _, m := range protogenMethods {
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

	generateBindBodyFunc(g, file)
	// output response
	generateOutputResponseFunc(g, file)

	generateRouterMethods(g, service, methods)
	generateHandlerMethods(g, service, methods)

	return nil
}

func createUniqueFile(plugin *protogen.Plugin, file *protogen.File) {
	if _, ok := uniqueFiles[string(file.GoImportPath)]; ok {
		return
	}
	uniqueFiles[string(file.GoImportPath)] = struct{}{}

	fileRouter := filepath.Dir(file.GeneratedFilenamePrefix) + "/" + filepath.Base(file.GeneratedFilenamePrefix) + ".s"
	plugin.NewGeneratedFile(fileRouter, file.GoImportPath)
}

func generateBindBodyFunc(g *protogen.GeneratedFile, file *protogen.File) {
	g.QualifiedGoIdent(protoIdent)
	g.QualifiedGoIdent(bindingJsonIdent)
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "io"})
	g.QualifiedGoIdent(bindingIdent)
	g.Import("unsafe")
	g.QualifiedGoIdent(protogen.GoIdent{GoImportPath: "errors"})

	fullname := string(file.Desc.FullName())

	g.P("const defaultMemory_", uniqueFileNames[fullname], " = 32 << 20")
	g.P("//go:linkname validate_", uniqueFileNames[fullname], " ", string(bindingIdent.GoImportPath), ".validate")
	g.P("func validate_", uniqueFileNames[fullname], "(obj any) error")
	g.P("//go:linkname mappingByPtr_", uniqueFileNames[fullname], " ", string(bindingIdent.GoImportPath), ".mappingByPtr")
	g.P("func mappingByPtr_", uniqueFileNames[fullname], "(ptr any, setter any, tag string) error")
	g.P()

	g.P("func _Bind_Gin_", uniqueFileNames[fullname], "(c *gin.Context, req proto.Message) error {")
	g.P("switch c.ContentType() {")
	g.P("case binding.MIMEMultipartPOSTForm:")
	{
		g.P("if err := c.Request.ParseMultipartForm(defaultMemory_", uniqueFileNames[fullname], "); err != nil {")
		g.P("return err")
		g.P("}")
		g.P("if err := mappingByPtr_", uniqueFileNames[fullname], "(req, c.Request, ", strconv.Quote("json"), "); err != nil {")
		g.P("return err")
		g.P("}")
	}
	g.P("case binding.MIMEPOSTForm:")
	{
		g.P("if err := c.Request.ParseForm(); err != nil {")
		g.P("return err")
		g.P("}")
		g.P("if err := c.Request.ParseMultipartForm(defaultMemory_", uniqueFileNames[fullname], "); err != nil && !errors.Is(err, http.ErrNotMultipart) {")
		g.P("return err")
		g.P("}")
		g.P("return binding.MapFormWithTag(req, c.Request.Form,", strconv.Quote("json"), ")")
	}
	g.P("default:")
	{
		g.P("bs, _ := io.ReadAll(c.Request.Body)")
		g.P("if len(bs) < 1 { return nil }")
		g.P("return (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(bs, req)")
	}
	g.P("}")
	g.P("return nil")
	g.P("}")
	g.P()
}

func generateOutputResponseFunc(g *protogen.GeneratedFile, file *protogen.File) {
	// output response
	fullname := string(file.Desc.FullName())
	g.P("func _Output_Gin_", uniqueFileNames[fullname], "(c *gin.Context, res any) {")
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
	fullname := string(service.Desc.ParentFile().FullName())
	g.P(fmt.Sprintf("func _%s_%s%d_Gin_Handler(srv %sGinServer) gin.HandlerFunc {", service.GoName, m.GoName, m.Num, service.GoName))
	// start handler
	{
		g.P("return func(c *gin.Context) {")
		g.P("req := new(", m.Input.GoIdent, ")")
		switch m.Input.Desc.FullName() {
		case "google.protobuf.Empty":
			break
		default:
			g.P("// param bind and validate")
			// Perform different bindings by request
			switch m.MethodName {
			case http.MethodGet, http.MethodDelete:
				generateBindQuery(g, len(m.Params) > 0, true)
			default:
				if len(m.Params) > 0 {
					generateBindParams(g, false)
				}
				g.P("if err := _Bind_Gin_", uniqueFileNames[fullname], "(c, req", m.Body, "); err != nil {")
				generateBindAbort(g)
				g.P("}")
			}
			// validate
			g.P("if err := validate_", uniqueFileNames[fullname], "(req); err != nil {")
			generateBindAbort(g)
			g.P("}")
		}

		// request response result
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
			g.P("_Output_Gin_", uniqueFileNames[fullname], "(c, res", resSuffix, ")")
		}
		g.P("}")
	}
	// end
	g.P("}")
	g.P()
}

func generateBindParams(g *protogen.GeneratedFile, checkErr bool) {
	g.P("// bind http.Request path params")
	g.P("{")
	{
		g.P("m := make(map[string][]string)")
		g.P("for _, v := range c.Params {")
		g.P("m[v.Key] = []string{v.Value}")
		g.P("}")
		if checkErr {
			g.P("if err := binding.MapFormWithTag(req, m, ", strconv.Quote("json"), "); err != nil {")
			generateBindAbort(g)
			g.P("}")
		} else {
			g.P("_ = binding.MapFormWithTag(req, m, ", strconv.Quote("json"), ")")
		}
	}
	g.P("}")
}

func generateBindQuery(g *protogen.GeneratedFile, hasParams bool, checkErr bool) {
	g.P("// bind http.Request query")
	g.P("query := c.Request.URL.Query()")
	if hasParams {
		g.P("for _, v := range c.Params {")
		g.P("if query.Get(v.Key) == ", strconv.Quote(""), "{")
		g.P("query.Set(v.Key, v.Value)")
		g.P("}")
		g.P("query[v.Key] = []string{v.Value}")
		g.P("}")
	}
	if checkErr {
		g.P("if err := binding.MapFormWithTag(req, query, ", strconv.Quote("json"), "); err != nil {")
		generateBindAbort(g)
		g.P("}")
	} else {
		g.P("_ = binding.MapFormWithTag(req, query, ", strconv.Quote("json"), ")")
	}
}

func generateBindAbort(g *protogen.GeneratedFile) {
	g.P("c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck")
	g.P("return")
}
