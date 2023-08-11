package camel

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Param struct {
	Name      string
	IsWild    bool
	OrignName string
}

// Parse 解析参数
func ParsePath(path string) []Param {
	if path == "" || path == "/" {
		return nil
	}
	pnames := []Param{}
	for i := 0; i < len(path); i++ {
		if path[i] != '{' {
			continue
		}
		pname := ""
		for j := i; j < len(path); j++ {
			switch path[j] {
			case '}':
				pname = path[i+1 : j]
				i = j
			case '/':
				j = len(path)
			case '\\':
				j = len(path)
			default:
			}
		}
		if pname != "" {
			p := Param{
				Name:      pname,
				OrignName: pname,
			}
			// 结构数据解析 参数名 = *|**|
			vs := strings.Split(pname, "=")
			if len(vs) == 2 {
				if vs[1] == "*" || vs[1] == "" {
					p.Name = vs[0]
				} else if strings.Trim(strings.TrimSpace(vs[1]), "*") == "" {
					p.Name = vs[0]
					p.IsWild = true
				}
			}
			pnames = append(pnames, p)
		}
	}
	return pnames
}

// FieldBindDescriptor 字段绑定描述符
func FieldBindDescriptor(field string, d protoreflect.Descriptor) protoreflect.Descriptor {
	if m, ok := d.(protoreflect.MessageDescriptor); ok {
		return m.Fields().ByName(protoreflect.Name(field))
	}
	return nil
}

// FieldsBindDescriptor 参数绑定描述符
func FieldsBindDescriptor(fields string, d protoreflect.Descriptor) protoreflect.Descriptor {
	fieldNames := strings.Split(fields, ".")
	var descriptor protoreflect.Descriptor
	for _, fieldName := range fieldNames {
		d = FieldBindDescriptor(fieldName, d)
		if d == nil {
			return nil
		}
		descriptor = d
	}
	return descriptor
}
