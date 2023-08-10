package frames

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cnartlu/protoc-gen-go-http/genhttp/camel"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func _replaceVarName(param camel.Param, varName string) string {
	if param.IsWild {
		return "*" + varName
	}
	return ":" + varName
}

type MethodDesc struct {
	*protogen.Method
	// Num方法格式
	Num int
	// Path 请求路由
	Path string
	// MethodName 请求方法
	MethodName string
	// Pname 路径参数名称
	Pnames []camel.Param
	// Params 参数
	Params map[string]protoreflect.Descriptor
	// Body字段值
	Body string
	// 选择器的其他HTTP绑定。嵌套绑定本身不能包含“additional_bindings”字段（即嵌套只能有一层深度）。
	ResponseBody string

	// replaceVarName 替换的变量名称
	replaceVarName func(param camel.Param, varName string) string
}

func (m MethodDesc) AddNum(i int) MethodDesc {
	m.Num += 1
	return m
}

func (m *MethodDesc) genParams() {
	m.Pnames = camel.ParsePath(m.Path)
	if len(m.Pnames) < 1 {
		return
	}
	// oldPath := m.Path
	m.Params = map[string]protoreflect.Descriptor{}
	for _, param := range m.Pnames {
		afterStr := param.Name
		if i := strings.LastIndex(param.Name, "."); i > 0 {
			afterStr = param.Name[i+1:]
		}
		m.Path = strings.Replace(m.Path, "{"+param.Name+"}", m.replaceVarName(param, afterStr), 1)
		descriptor := camel.FieldsBindDescriptor(param.Name, m.Desc.Input())
		if descriptor == nil {
			// fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field [%s] in path:'%s' not found in [%s].\n", param, oldPath, m.Input.Desc.FullName())
			continue
		}
		m.Params[afterStr] = descriptor
	}
}

func NewMethodDesc(m *protogen.Method, rule *annotations.HttpRule) MethodDesc {
	md := MethodDesc{
		Method:         m,
		replaceVarName: _replaceVarName,
	}
	if rule == nil {
		service := m.Parent
		path := fmt.Sprintf("/%s/%s", service.Desc.FullName(), m.Desc.Name())
		md.Path = path
		md.MethodName = http.MethodPost
		return md
	}

	var (
		path   string
		method string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = http.MethodGet
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = http.MethodPut
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = http.MethodPost
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = http.MethodDelete
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = http.MethodPatch
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	md.Path = path
	md.MethodName = method

	switch rule.Body {
	case "", "*":
		md.Body = ""
	default:
		md.Body = "." + camel.CaseVars(rule.Body)
	}

	switch rule.ResponseBody {
	case "", "*":
		md.ResponseBody = ""
	default:
		md.ResponseBody = "." + camel.CaseVars(rule.ResponseBody)
	}

	md.genParams()

	return md
}
