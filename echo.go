package main

var echoTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

// {{$svrType}}HttpServer is the server API for {{$svrType}} service.
// All implementations must embed Unimplemented{{$svrType}}HttpServer
// for forward compatibility
type {{$svrType}}HttpServer interface {
	{{$svrType}}Server
	// Bind(v4.Context, any) error
	{{- range .Methods}}
	Bind{{.RequestBindName}}(v4.Context) (*{{.Request}}, error)
	{{- end}}
	mustEmbedUnimplemented{{$svrType}}HttpServer()
}

// Unimplemented{{$svrType}}HttpServer must be embedded to have forward compatible implementations.
type Unimplemented{{$svrType}}HttpServer struct {
}

func (Unimplemented{{$svrType}}HttpServer) Bind(c v4.Context, v any) error {
	return v4.NewHTTPError(http.StatusBadRequest, "bind method Bind not implemented")
}
{{- range .Methods}}
func (Unimplemented{{$svrType}}HttpServer) Bind{{.RequestBindName}}(c v4.Context) (*{{.Request}}, error) {
	return nil, v4.NewHTTPError(http.StatusBadRequest, "bind method Bind{{.RequestBindName}} not implemented")
}
{{- end}}
func (Unimplemented{{$svrType}}HttpServer) mustEmbedUnimplemented{{$svrType}}HttpServer() {}

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
		if req, err = srv.Bind{{.RequestBindName}}(c); err != nil {
			return err
		}
		if err := c.Validate(req); err != nil {
			return err
		}
		if res, err = srv.{{.Name}}(c.Request().Context(), req); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}
{{end}}

// {{range .Methods}}
// func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler_0(srv {{$svrType}}Server) v4.HandlerFunc {
// 	return func(c v4.Context) error {
// 		var (
// 			req *{{.Request}} = new({{.Request}})
// 			res *{{.Reply}} = new({{.Reply}})
// 			err error
// 			ctx = c.Request().Context()
// 			coding encoding.Codec
// 			unmarshalData []byte
// 		)
// 		{{- if eq .Method "GET"}}
// 		coding = encoding.GetCodec(form.Name)
// 		urlValues := c.Request().URL.Query()
// 		{{- if .HasVars}}
// 		{{- range $k, $v := .Vars}}
// 		if !urlValues.Has("{{$k}}") {
// 			urlValues.Set("{{$k}}", c.Param("{{$k}}"))
// 		}
// 		{{- end}}
// 		{{- end}}
// 		unmarshalData = []byte(urlValues.Encode())
// 		{{- else}}
// 		ctype := c.Request().Header.Get(v4.HeaderContentType)
// 		switch {
// 		case strings.HasPrefix(ctype, v4.MIMEApplicationJSON):
// 			unmarshalData, err = io.ReadAll(c.Request().Body)
// 			if err != nil {
// 				return err
// 			}
// 			unmarshalDataLen := len(unmarshalData)
// 			if unmarshalDataLen < 1 {
// 				unmarshalData = []byte("{}")
// 			}
// 			{{- if .HasVars}}
// 			jsonStr := ""
// 			{{- range $k, $v := .Vars}}
// 			jsonStr = jsonStr+ "\"{{$k}}\":\"" + c.Param("{{$k}}") + "\","
// 			{{- end}}
// 			afterByte := append([]byte(jsonStr[:len(jsonStr)-1]), unmarshalData[1:]...)
// 			unmarshalData = append(unmarshalData[:1], afterByte...)
// 			{{- end}}
// 			coding = encoding.GetCodec(json.Name)
// 		case strings.HasPrefix(ctype, v4.MIMEApplicationXML), strings.HasPrefix(ctype, v4.MIMETextXML):
// 			unmarshalData, err = io.ReadAll(c.Request().Body)
// 			if err != nil {
// 				return err
// 			}
// 			coding = encoding.GetCodec(xml.Name)
// 		case strings.HasPrefix(ctype, v4.MIMEApplicationForm), strings.HasPrefix(ctype, v4.MIMEMultipartForm):
// 			params, err := c.FormParams()
// 			if err != nil {
// 				return v4.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
// 			}
// 			urlValues := params
// 			{{- if .HasVars}}
// 			{{- range $k, $v := .Vars}}
// 			if !urlValues.Has("{{$k}}") {
// 				urlValues.Set("{{$k}}", c.Param("{{$k}}"))
// 			}
// 			{{- end}}
// 			{{- end}}
// 			unmarshalData = []byte(urlValues.Encode())
// 			coding = encoding.GetCodec(form.Name)
// 		default:
// 			return v4.ErrUnsupportedMediaType
// 		}
// 		{{- end}}
// 		if coding == nil {
// 			err = c.Bind(req)
// 		} else {
// 			err = coding.Unmarshal(unmarshalData, req)
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		if res, err = srv.{{.Name}}(context.WithValue(ctx, v4.DefaultBinder{}, c), req); err != nil {
// 			return err
// 		}
// 		return c.JSON(http.StatusOK, res)
// 	}
// }
// {{end}}
`
