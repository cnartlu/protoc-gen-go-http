package main

var echoTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{$svrType}}ServiceRegistrar interface {
	Add(method, path string, handler v4.HandlerFunc, middleware ...v4.MiddlewareFunc) *v4.Route
}

func Register{{.ServiceType}}HttpServer(r {{$svrType}}ServiceRegistrar, srv {{$svrType}}Server) {
	{{- range .Methods}}
	r.Add("{{.Method}}", "{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}Server) v4.HandlerFunc {
	return func(c v4.Context) error {
		var req *{{.Request}} = new({{.Request}})
		var err error
		ctx := c.Request().Context()
		var coding encoding.Codec
		var unmarshalData []byte
		{{- if eq .Method "GET"}}
		coding = encoding.GetCodec(form.Name)
		urlValues := c.Request().URL.Query()
		{{- if .HasVars}}
		{{- range $k, $v := .Vars}}
		if !urlValues.Has("{{$k}}") {
			urlValues.Set("{{$k}}", c.Param("{{$k}}"))
		}
		{{- end}}
		{{- end}}
		unmarshalData = []byte(urlValues.Encode())
		{{- else}}
		ctype := c.Request().Header.Get(v4.HeaderContentType)
		switch {
		case strings.HasPrefix(ctype, v4.MIMEApplicationJSON):
			unmarshalData, err = io.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}
			unmarshalDataLen := len(unmarshalData)
			if unmarshalDataLen < 1 {
				unmarshalData = []byte("{}")
			}
			{{- if .HasVars}}
			jsonStr := ""
			{{- range $k, $v := .Vars}}
			jsonStr = jsonStr+ "\"{{$k}}\":\"" + c.Param("{{$k}}") + "\","
			{{- end}}
			afterByte := append([]byte(jsonStr[:len(jsonStr)-1]), unmarshalData[1:]...)
			unmarshalData = append(unmarshalData[:1], afterByte...)
			{{- end}}
			coding = encoding.GetCodec(json.Name)
		case strings.HasPrefix(ctype, v4.MIMEApplicationXML), strings.HasPrefix(ctype, v4.MIMETextXML):
		case strings.HasPrefix(ctype, v4.MIMEApplicationForm), strings.HasPrefix(ctype, v4.MIMEMultipartForm):
			params, err := c.FormParams()
			if err != nil {
				return v4.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
			}
			urlValues := params
			{{- if .HasVars}}
			{{- range $k, $v := .Vars}}
			if !urlValues.Has("{{$k}}") {
				urlValues.Set("{{$k}}", c.Param("{{$k}}"))
			}
			{{- end}}
			{{- end}}
			unmarshalData = []byte(urlValues.Encode())
			coding = encoding.GetCodec(form.Name)
		default:
			return v4.ErrUnsupportedMediaType
		}
		{{- end}}
		if coding == nil {
			err = c.Bind(req)
		} else {
			err = coding.Unmarshal(unmarshalData, req)
		}
		if err != nil {
			return err
		}
		if err := c.Validate(req{{.Body}}); err != nil && err != v4.ErrValidatorNotRegistered {
			return err
		}
		res, err := srv.{{.Name}}(context.WithValue(ctx, v4.DefaultBinder{}, c), req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	}
}
{{end}}
`
