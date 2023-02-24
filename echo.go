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
		ctx := c.Request().Context()
		if err := c.Bind(req{{.Body}}); err != nil {
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

var echoMessageTemplate = `

`
