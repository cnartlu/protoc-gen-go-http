package gin

var ginTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

func Register{{.ServiceType}}HttpServer(r gin.IRouter, srv {{$svrType}}Server, binding {{$svrType}}Binding) {
	_{{$svrType}}_ServiceHttpDesc = srv
	_{{$svrType}}_ServiceBinding.binding = binding
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler)
	{{- end}}
}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(c *gin.Context) {
	var req *{{.Request}} = new({{.Request}})
	ctx := c.Request.Context()
	if err := _{{$svrType}}_ServiceBinding.Bind(c, req); err != nil {
		c.Abort()
		c.Error(err).SetType(gin.ErrorTypeBind)
		return 
	}
	res, err := _{{$svrType}}_ServiceHttpDesc.{{.Name}}(context.WithValue(ctx, gin.ContextKey, c), req)
	if err != nil {
		c.Abort()
		c.Error(err)
		return 
	}
	c.JSON(http.StatusOK, res)
}
{{end}}

const (
	_{{$svrType}}MaxMemory = 32 << 20 // 32 MB
	_{{$svrType}}DefaultTag = "json"
)

type {{$svrType}}Binding interface{
	Bind(c *gin.Context, v interface{}) error
}

type _{{$svrType}}Binding struct {
	binding {{$svrType}}Binding
}

func (b _{{$svrType}}Binding) Bind(c *gin.Context, v interface{}) error {
	if b.binding != nil {
		return b.binding.Bind(c, v)
	}
	if len(c.Params) > 0 {
		m := make(map[string][]string)
		for _, v := range c.Params {
			m[v.Key] = []string{v.Value}
		}
		if err := binding.MapFormWithTag(v, m, _{{$svrType}}DefaultTag); err != nil {
			return err
		}
	}
	req := c.Request
	if req.Method == http.MethodGet {
		values := req.URL.Query()
		if err := binding.MapFormWithTag(v, values, _{{$svrType}}DefaultTag); err != nil {
			return err
		}
	}
	useBinding := binding.Default(req.Method, c.ContentType())
	switch useBinding{
	case binding.Form:
		if err := req.ParseForm(); err != nil {
			return err
		}
		if err := req.ParseMultipartForm(_AccountMaxMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
			return err
		}
		return binding.MapFormWithTag(v, req.Form, _{{$svrType}}DefaultTag)
	case binding.FormMultipart:
		fallthrough
	default:
		return useBinding.Bind(req, v)
	}
}

var (
	_{{$svrType}}_ServiceHttpDesc {{$svrType}}Server = Unimplemented{{$svrType}}Server{}
	_{{$svrType}}_ServiceBinding _{{$svrType}}Binding = _{{$svrType}}Binding{}
)
`
