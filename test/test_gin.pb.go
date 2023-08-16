// Code generated by protoc-plugin-http. DO NOT EDIT.
// versions:
// - protoc-plugin-http v0.1.1
// - protoc  v3.19.4
// source: test/test.proto

package test

import (
	context "context"
	errors "errors"
	gin "github.com/gin-gonic/gin"
	binding "github.com/gin-gonic/gin/binding"
	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	io "io"
	http "net/http"
	_ "unsafe"
)

// defaultMemory default maximum parsing memory
const defaultMemory = 32 << 20

//go:linkname validate github.com/gin-gonic/gin/binding.validate
func validate(obj any) error

//go:linkname mappingByPtr github.com/gin-gonic/gin/binding.mappingByPtr
func mappingByPtr(ptr any, setter any, tag string) error

//go:linkname ginParseAccept github.com/gin-gonic/gin.parseAccept
func ginParseAccept(acceptHeader string) []string

// RequestGinHandler customize the binding function, the binding method can be determined by binding parameter type and context
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
		case "application/xml", "text/xml":
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
}

func _Bind_Gin_Params(c *gin.Context, req proto.Message) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.MapFormWithTag(req, m, BindGinTagName)
}

// _Must_Bind_Gin_Params must bind params
func _Must_Bind_Gin_Params(c *gin.Context, req proto.Message) error {
	if err := _Bind_Gin_Params(c, req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

func _Bind_Gin_Query(c *gin.Context, req proto.Message) error {
	query := c.Request.URL.Query()
	for _, v := range c.Params {
		if query.Get(v.Key) == "" {
			query.Set(v.Key, v.Value)
		}
	}
	return binding.MapFormWithTag(req, query, BindGinTagName)
}

// _Must_Bind_Gin_Query must bind query
func _Must_Bind_Gin_Query(c *gin.Context, req proto.Message) error {
	if err := _Bind_Gin_Query(c, req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
		return err
	}
	return nil
}

// TestGinServer is the server API for Test service.
// All implementations must embed UnimplementedTestGinServer
// for forward compatibility
type TestGinServer interface {
	List(ctx context.Context, req *ListTestRequest) (*ListTestReply, error)
	Get(ctx context.Context, req *GetTestRequest) (*GetTestReply, error)
	Create(ctx context.Context, req *CreateTestRequest) (*CreateTestReply, error)
	Update(ctx context.Context, req *UpdateTestRequest) (*UpdateTestReply, error)
	Delete(ctx context.Context, req *DeleteTestRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestGinServer must be embedded to have forward compatible implementations.
type UnimplementedTestGinServer struct{}

func (UnimplementedTestGinServer) List(ctx context.Context, req *ListTestRequest) (*ListTestReply, error) {
	return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}
}
func (UnimplementedTestGinServer) Get(ctx context.Context, req *GetTestRequest) (*GetTestReply, error) {
	return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}
}
func (UnimplementedTestGinServer) Create(ctx context.Context, req *CreateTestRequest) (*CreateTestReply, error) {
	return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}
}
func (UnimplementedTestGinServer) Update(ctx context.Context, req *UpdateTestRequest) (*UpdateTestReply, error) {
	return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}
}
func (UnimplementedTestGinServer) Delete(ctx context.Context, req *DeleteTestRequest) (*emptypb.Empty, error) {
	return nil, gin.Error{Type: gin.ErrorTypePublic, Err: errors.New(http.StatusText(http.StatusNotImplemented))}
}
func (UnimplementedTestGinServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestGinServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestGinServer will
// result in compilation errors.
type UnsafeTestGinServer interface {
	mustEmbedUnimplementedTestServer()
}

type TestGinRouter = gin.IRoutes

func RegisterTestGinServer(r TestGinRouter, srv TestGinServer) {
	r.GET("/account/:id/:c_asa_2c_3/:kk/*aa", _Test_List0_Gin_Handler(srv))
	r.POST("/as/a/::id", _Test_List1_Gin_Handler(srv))
	r.GET("/account/:id", _Test_Get0_Gin_Handler(srv))
	r.POST("/account", _Test_Create0_Gin_Handler(srv))
	r.PUT("/account/:id", _Test_Update0_Gin_Handler(srv))
	r.DELETE("/account", _Test_Delete0_Gin_Handler(srv))
}

func _Test_List0_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(ListTestRequest)
		if err := _Must_Bind_Gin_Params(c, req); err != nil {
			return
		}
		if err := _Must_Bind_Gin_Query(c, req); err != nil {
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.List(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}

func _Test_List1_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(ListTestRequest)
		if err := _Must_Bind_Gin_Params(c, req); err != nil {
			return
		}
		if err := bindGinRequestBodyHandler(c, req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.List(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}

func _Test_Get0_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(GetTestRequest)
		if err := _Must_Bind_Gin_Query(c, req); err != nil {
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.Get(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}

func _Test_Create0_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(CreateTestRequest)
		if err := bindGinRequestBodyHandler(c, req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.Create(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}

func _Test_Update0_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(UpdateTestRequest)
		if err := bindGinRequestBodyHandler(c, req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.Update(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}

func _Test_Delete0_Gin_Handler(srv TestGinServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := new(DeleteTestRequest)
		if err := _Must_Bind_Gin_Query(c, req); err != nil {
			return
		}
		if err := validate(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err).SetType(gin.ErrorTypeBind) //nolint: errcheck
			return
		}
		res, err := srv.Delete(c, req)
		if err != nil {
			c.Abort()
			c.Error(err)
			return
		}
		outGinResponseHandler(c, res)
	}
}
