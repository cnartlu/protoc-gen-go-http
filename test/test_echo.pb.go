// Code generated by protoc-plugin-http. DO NOT EDIT.
// versions:
// - protoc-plugin-http v0.1.1
// - protoc  v3.19.4
// source: test/test.proto

package test

import (
	context "context"
	v4 "github.com/labstack/echo/v4"
	proto "google.golang.org/protobuf/proto"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	http "net/http"
	strings "strings"
)

func echoParseAccept(acceptHeader string) []string {
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
}

func _OutEchoResponseHandler(c v4.Context, res any) error {
	accepted := echoParseAccept(c.Request().Header.Get("Accept"))
	for _, accept := range accepted {
		switch accept {
		case "application/json":
			return c.JSON(http.StatusOK, res)
		case "application/xml", "text/xml":
			return c.XML(http.StatusOK, res)
		case "application/x-protobuf", "application/protobuf":
			bs, _ := proto.Marshal(res.(proto.Message))
			return c.Blob(http.StatusOK, "application/x-protobuf", bs)
		default:
		}
	}
	return c.JSON(http.StatusOK, res)
}

// TestEchoServer is the server API for Test service.
// All implementations must embed UnimplementedTestEchoServer
// for forward compatibility
type TestEchoServer interface {
	List(ctx context.Context, req *ListTestRequest) (*ListTestReply, error)
	Get(ctx context.Context, req *GetTestRequest) (*GetTestReply, error)
	Create(ctx context.Context, req *CreateTestRequest) (*CreateTestReply, error)
	Update(ctx context.Context, req *UpdateTestRequest) (*UpdateTestReply, error)
	Delete(ctx context.Context, req *DeleteTestRequest) (*emptypb.Empty, error)
	mustEmbedUnimplementedTestServer()
}

// UnimplementedTestEchoServer must be embedded to have forward compatible implementations.
type UnimplementedTestEchoServer struct {
}

func (UnimplementedTestEchoServer) List(ctx context.Context, req *ListTestRequest) (*ListTestReply, error) {
	return nil, v4.ErrNotImplemented
}
func (UnimplementedTestEchoServer) Get(ctx context.Context, req *GetTestRequest) (*GetTestReply, error) {
	return nil, v4.ErrNotImplemented
}
func (UnimplementedTestEchoServer) Create(ctx context.Context, req *CreateTestRequest) (*CreateTestReply, error) {
	return nil, v4.ErrNotImplemented
}
func (UnimplementedTestEchoServer) Update(ctx context.Context, req *UpdateTestRequest) (*UpdateTestReply, error) {
	return nil, v4.ErrNotImplemented
}
func (UnimplementedTestEchoServer) Delete(ctx context.Context, req *DeleteTestRequest) (*emptypb.Empty, error) {
	return nil, v4.ErrNotImplemented
}
func (UnimplementedTestEchoServer) mustEmbedUnimplementedTestServer() {}

// UnsafeTestEchoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TestEchoServer will
// result in compilation errors.
type UnsafeTestEchoServer interface {
	mustEmbedUnimplementedTestServer()
}

type TestHttpRouter interface {
	Add(method, path string, handler v4.HandlerFunc, middleware ...v4.MiddlewareFunc) *v4.Route
}

func RegisterTestEchoServer(r TestHttpRouter, srv TestEchoServer) {
	r.Add("GET", "/account/:id/:c_asa_2c_3/:kk/*aa", _Test_List0_Echo_Handler(srv))
	r.Add("POST", "/as/a/::id", _Test_List1_Echo_Handler(srv))
	r.Add("GET", "/account/:id", _Test_Get0_Echo_Handler(srv))
	r.Add("POST", "/account", _Test_Create0_Echo_Handler(srv))
	r.Add("PUT", "/account/:id", _Test_Update0_Echo_Handler(srv))
	r.Add("DELETE", "/account", _Test_Delete0_Echo_Handler(srv))
}

func _Test_List0_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(ListTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.List(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}

func _Test_List1_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(ListTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.List(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}

func _Test_Get0_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(GetTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.Get(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}

func _Test_Create0_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(CreateTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.Create(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}

func _Test_Update0_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(UpdateTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.Update(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}

func _Test_Delete0_Echo_Handler(srv TestEchoServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		req := new(DeleteTestRequest)
		if err := c.Bind(req); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.Delete(c.Request().Context(), req)
		if err != nil {
			return err
		}
		return _OutEchoResponseHandler(c, res)
	}
}
