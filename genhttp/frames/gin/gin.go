package gin

import (
	"fmt"
	"net/http"
	"path/filepath"

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
	importBinding   = protogen.GoImportPath("github.com/gin-gonic/gin/binding")
	importProtoJson = protogen.GoImportPath("google.golang.org/protobuf/encoding/protojson")
	importProto     = protogen.GoImportPath("google.golang.org/protobuf/proto")
	importNetHttp   = protogen.GoImportPath("net/http")
	importGin       = protogen.GoImportPath("github.com/gin-gonic/gin")
	importErrors    = protogen.GoImportPath("errors")
	importContext   = protogen.GoImportPath("context")
	importUnsafe    = protogen.GoImportPath("unsafe")
	importIo        = protogen.GoImportPath("io")
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

	g.QualifiedGoIdent(importBinding.Ident(""))
	g.QualifiedGoIdent(importIo.Ident(""))
	g.QualifiedGoIdent(importProtoJson.Ident(""))
	g.QualifiedGoIdent(importGin.Ident(""))
	g.Import(importUnsafe)

	// Add an assembly file with the suffix name s(.s) to ensure that the go build command is executable,
	// for details https://pkg.go.dev/cmd/compile
	fileRouter := filepath.Dir(file.GeneratedFilenamePrefix) + "/" + filepath.Base(file.GeneratedFilenamePrefix) + ".s"
	plugin.NewGeneratedFile(fileRouter, file.GoImportPath)

	g.P("// defaultMemory default maximum parsing memory")
	g.P("const defaultMemory", " = 32 << 20")
	g.P()
	//
	g.P("//go:linkname validate ", string(importBinding), ".validate")
	g.P("func validate(obj any) error")
	g.P()
	//
	g.P("//go:linkname mappingByPtr ", string(importBinding), ".mappingByPtr")
	g.P("func mappingByPtr(ptr any, setter any, tag string) error")
	g.P()
	//
	g.P("//go:linkname ginParseAccept ", string(importGin), ".parseAccept")
	g.P("func ginParseAccept(acceptHeader string) []string")
	g.P()
	//
	g.P(`// RequestGinHandler customize the binding function, the binding method can be determined by binding parameter type and context
	type RequestGinHandler func(c *gin.Context, req proto.Message) error
	// ResponseGinHandler Custom response function, output response
	type ResponseGinHandler func(c *gin.Context, res any)

	var (
		BindGinTagName string = "json"
		// _BindGinRequestHandler binding handler
		_BindGinRequestHandler RequestGinHandler
		// _OutGinResponseHandler output response handler
		_OutGinResponseHandler ResponseGinHandler

		JSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	)

	func bindGinRequestBodyHandler(c *gin.Context, req proto.Message) error {
		if _BindGinRequestHandler != nil {
			return _BindGinRequestHandler(c, req)
		}
		switch c.ContentType() {
		case binding.MIMEMultipartPOSTForm:
			if err := c.Request.ParseMultipartForm(defaultMemory); err != nil {
				return err
			}
			if err := mappingByPtr(req, c.Request, BindGinTagName); err != nil {
				return err
			}
		case binding.MIMEPOSTForm:
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
			if err := c.Request.ParseMultipartForm(defaultMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
				return err
			}
			return binding.MapFormWithTag(req, c.Request.Form, BindGinTagName)
		default:
			bs, _ := io.ReadAll(c.Request.Body)
			if len(bs) < 1 {
				return nil
			}
			return JSON.Unmarshal(bs, req)
		}
		return nil
	}

	func outGinResponseHandler(c *gin.Context, res any) {
		if c.Accepted == nil {
			c.Accepted = ginParseAccept(c.Request.Header.Get("Accept"))
		}
		if _OutGinResponseHandler != nil {
			_OutGinResponseHandler(c, res)
			return
		}
		for _, accept := range c.Accepted {
			switch accept {
			case "application/json":
				c.JSON(http.StatusOK, res)
				return
			case "application/xml","text/xml":
				c.XML(http.StatusOK, res)
				return
			case "application/x-protobuf", "application/protobuf":
				c.ProtoBuf(http.StatusOK, res)
				return
			default:
			}
		}
		c.JSON(http.StatusOK, res)
	}

	// SetBindGinRequestHandler set binding handler
	func SetBindGinRequestHandler(v RequestGinHandler) {
		_BindGinRequestHandler = v
	}
	
	// SetOutGinResponseHandler set output response handler
	func SetOutGinResponseHandler(v ResponseGinHandler) {
		_OutGinResponseHandler = v
	}`)
	g.P()

	generateBindParamsFunc(g)
	generateBindQueryFunc(g)
}

func generateBindParamsFunc(g *protogen.GeneratedFile) {
	ginContext := g.QualifiedGoIdent(importGin.Ident("Context"))
	ginErrorTypeBind := g.QualifiedGoIdent(importGin.Ident("ErrorTypeBind"))
	protoMessage := g.QualifiedGoIdent(importProto.Ident("Message"))
	httpStatusBadRequest := g.QualifiedGoIdent(importNetHttp.Ident("StatusBadRequest"))
	g.P("func _Bind_Gin_Params(c *", ginContext, ", req ", protoMessage, ") error {")
	g.P(`m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.MapFormWithTag(req, m, BindGinTagName)`)
	g.P("}")
	g.P()
	//
	g.P("// _Must_Bind_Gin_Params must bind params")
	g.P("func _Must_Bind_Gin_Params(c *", ginContext, ", req ", protoMessage, ") error {")
	g.P("if err := _Bind_Gin_Params(c, req); err != nil {")
	g.P("c.AbortWithError(", httpStatusBadRequest, ", err).SetType(", ginErrorTypeBind, ") //nolint: errcheck")
	g.P("return err")
	g.P("}")
	g.P("return nil")
	g.P("}")
	g.P()
}

func generateBindQueryFunc(g *protogen.GeneratedFile) {
	ginContext := g.QualifiedGoIdent(importGin.Ident("Context"))
	ginErrorTypeBind := g.QualifiedGoIdent(importGin.Ident("ErrorTypeBind"))
	protoMessage := g.QualifiedGoIdent(importProto.Ident("Message"))
	httpStatusBadRequest := g.QualifiedGoIdent(importNetHttp.Ident("StatusBadRequest"))
	g.P("func _Bind_Gin_Query(c *", ginContext, ", req ", protoMessage, ") error {")
	g.P(`query := c.Request.URL.Query()
		for _, v := range c.Params {
			if query.Get(v.Key) == "" {
				query.Set(v.Key, v.Value)
			}
		}
		return binding.MapFormWithTag(req, query, BindGinTagName)`)
	g.P("}")
	g.P()
	//
	g.P("// _Must_Bind_Gin_Query must bind query")
	g.P("func _Must_Bind_Gin_Query(c *", ginContext, ", req ", protoMessage, ") error {")
	g.P("if err := _Bind_Gin_Query(c, req); err != nil {")
	g.P("c.AbortWithError(", httpStatusBadRequest, ", err).SetType(", ginErrorTypeBind, ") //nolint: errcheck")
	g.P("return err")
	g.P("}")
	g.P("return nil")
	g.P("}")
	g.P()
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
	g.P("return func(c *", ginContextIdent, ") {")
	g.P("req := new(", m.Input.GoIdent, ")")
	generateBindRequest(g, m)
	g.P("res, err := srv.", m.GoName, "(c, req)")
	g.P("if err != nil { c.Abort() \n c.Error(err) \n return }")
	if m.ResponseBody != "" {
		g.P("outGinResponseHandler(c, res", m.ResponseBody, ")")
	} else {
		g.P("outGinResponseHandler(c, res)")
	}
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
		g.P(`if err := _Must_Bind_Gin_Params(c, req); err != nil {
			return
		}`)
	}

	abortErrorStr := fmt.Sprintf("c.AbortWithError(%s, err).SetType(%s) //nolint: errcheck", httpStatusBadRequest, ginErrorTypeBind)

	switch m.MethodName {
	case http.MethodGet, http.MethodDelete:
		g.P(`if err := _Must_Bind_Gin_Query(c, req`, m.Body, `); err != nil {
			return
		}`)
	default:
		g.P("if err := bindGinRequestBodyHandler(c, req", m.Body, "); err != nil {")
		g.P(abortErrorStr)
		g.P("return")
		g.P("}")
	}

	g.P("if err := validate(req); err != nil {")
	g.P(abortErrorStr)
	g.P("return")
	g.P("}")
}
