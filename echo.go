package main

var echoTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

// {{$svrType}}HttpServer is the server API for {{$svrType}} service.
// All implementations must embed Unimplemented{{$svrType}}HttpServer
// for forward compatibility
type {{$svrType}}HttpServer interface {
	{{$svrType}}Server
	SetHttpResponse(v4.Context, any) error
	{{- range .Methods}}
	{{- if not .RequestBindHide}}
	Bind{{.RequestBindName}}(v4.Context) (*{{.Request}}, error)
	{{- end}}
	{{- end}}
	mustEmbedUnimplemented{{$svrType}}HttpServer()
}

// Unimplemented{{$svrType}}HttpServer must be embedded to have forward compatible implementations.
type Unimplemented{{$svrType}}HttpServer struct {
}

func (u Unimplemented{{$svrType}}HttpServer) SetHttpResponse(c v4.Context, v any) error {
	return c.JSON(http.StatusOK, v)
}
{{- range .Methods}}
	{{- if not .RequestBindHide}}
		func (u Unimplemented{{$svrType}}HttpServer) Bind{{.RequestBindName}}(c v4.Context) (*{{.Request}}, error) {
			{{- if gt .Num 0}}
			return u.Bind{{.RequestBindOriginName}}(c)
		{{- else}}
			var req *{{.Request}} = new({{.Request}})
			{{- if eq .Method "GET"}}
			if err := u.bindByQuery(c, req); err != nil {
				return nil, err
			}
			{{- else}}
			if err := u.bindByJson(c, req); err != nil {
				return nil, err
			}
			{{- end}}
			{{- /* 绑定路径参数 */ -}}
			{{- if .HasVars}}
				// TODO: 未实现参数绑定，后续应当支持
				{{- range $k, $v := .Vars}}
				// if v := c.Param("{{$k}}"); v != "" {
				// 	d, err := strcovc.ParseInt(c.Param("{{$k}}"), 10, 64)
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	req.{{$v}} = uint64(d)
				// }
				{{- end}}
			{{- end}}
			return req, nil
		{{- end}}
		}
	{{- end}}
{{- end}}
func (u Unimplemented{{$svrType}}HttpServer) bindByQuery(c v4.Context, v any) error {
	coding := encoding.GetCodec(form.Name)
	if err := coding.Unmarshal([]byte(c.Request().URL.Query().Encode()), v); err != nil {
		return err
	}
	return nil
}
func (u Unimplemented{{$svrType}}HttpServer) bindByJson(c v4.Context, v any) error {
	if c.Request().ContentLength > 0 {
		coding := encoding.GetCodec(json.Name)
		r := io.LimitReader(c.Request().Body, 32 << 20)
		b, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		if err := coding.Unmarshal(b, v); err != nil {
			return err
		}
	}
	return nil
}
func (Unimplemented{{$svrType}}HttpServer) mustEmbedUnimplemented{{$svrType}}HttpServer() {
	{{- /* return nil, v4.NewHTTPError(http.StatusBadRequest, "bind method Bind not implemented") */ -}}
}

// Unsafe{{$svrType}}HttpServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to {{$svrType}}HttpServer will
// result in compilation errors.
type Unsafe{{$svrType}}HttpServer interface {
	mustEmbedUnimplemented{{$svrType}}HttpServer()
}

type {{$svrType}}ServiceRegistrar interface {
	Add(method, path string, handler v4.HandlerFunc, middleware ...v4.MiddlewareFunc) *v4.Route
}

func Register{{.ServiceType}}HttpServer(r {{$svrType}}ServiceRegistrar, srv {{$svrType}}HttpServer) {
	{{- range .Methods}}
	r.Add("{{.Method}}", "{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HttpServer) v4.HandlerFunc {
	return func(c v4.Context) error {
		var (
			req *{{.Request}} = new({{.Request}})
			res *{{.Reply}} = new({{.Reply}})
			err error
		)
		{{- if not .RequestBindHide}}
		if req, err = srv.Bind{{.RequestBindName}}(c); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		{{- end}}
		if res, err = srv.{{.Name}}(c.Request().Context(), req); err != nil {
			return err
		}
		return srv.SetHttpResponse(c, res)
	}
}
{{end}}
`
