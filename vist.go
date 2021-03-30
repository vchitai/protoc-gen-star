package pgs

import (
	"fmt"
	"strings"

	_ "github.com/gogo/protobuf/gogoproto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Function struct {
	Method string
	Path   string
	Name   string
	Extra  string
}

type ExtensibleFile struct {
	File
}

func (ef ExtensibleFile) AddMethod(target string, function Function) {
	for _, s := range ef.Services() {
		for _, m := range s.Methods() {
			if m.Name().String() == target {
				return
			}
		}
		name := fmt.Sprintf("%s", function.Name)
		in := fmt.Sprintf("%sRequest", function.Name)
		out := fmt.Sprintf("%sResponse", function.Name)
		opt := fmt.Sprintf("[google.api.http]:{%s:\"%s\" %s}\n", function.Method, function.Path, function.Extra)
		x := new(descriptorpb.MethodOptions)
		prototext.Unmarshal([]byte(opt), x)
		s.addMethod(&method{
			desc: &descriptorpb.MethodDescriptorProto{
				Name:       &name,
				InputType:  &in,
				OutputType: &out,
				Options:    x,
			},
			in: &msg{
				desc: &descriptorpb.DescriptorProto{Name: &in},
			},
			out: &msg{
				desc: &descriptorpb.DescriptorProto{Name: &out},
			},
			service: s,
			options: fmt.Sprintf("[google.api.http]:{%s:\"%s\" %s}\n", function.Method, function.Path, function.Extra),
		})
	}
}
func (ef ExtensibleFile) AddMessage(name string) {
	for _, m := range ef.Messages() {
		if m.Name().String() == name {
			return
		}
	}
	ef.addMessage(&msg{
		desc: &descriptorpb.DescriptorProto{
			Name: &name,
		},
	})
}

type DescriberMixin interface {
	DescribeSelf() string
}

var _ DescriberMixin = &XFile{}
var _ DescriberMixin = &XField{}
var _ DescriberMixin = &XEnum{}
var _ DescriberMixin = &XEnumValue{}
var _ DescriberMixin = &XMessage{}
var _ DescriberMixin = &XMethod{}
var _ DescriberMixin = &XOneOf{}
var _ DescriberMixin = &XPackage{}
var _ DescriberMixin = &XService{}

type XFile struct {
	File
}

func (x XFile) DescribeSelf() string {
	imp := ""
	for _, i := range x.Imports() {
		imp += fmt.Sprintf("import \"%s\";\n", i.Name().String())
	}
	s := ""

	for _, service := range x.Services() {
		s += XService{service}.DescribeSelf()
	}

	for _, m := range x.Messages() {
		s += XMessage{m}.DescribeSelf()
	}

	for _, e := range x.Enums() {
		s += XEnum{e}.DescribeSelf()
	}

	return fmt.Sprintf(
		`
syntax = "%s";
package %s;

option (gogoproto.unmarshaler_all) = true;
option (gogoproto.sizer_all) = true;
option (gogoproto.equal_all) = true;
option (gogoproto.marshaler_all) = true;

%s
%s
`, x.Syntax().String(), x.Package().ProtoName().String(), imp, s,
	)
}

type XPackage struct {
	Package
}

func (x XPackage) DescribeSelf() string {
	return fmt.Sprintf("package %s", x.ProtoName().String())
}

type XService struct {
	Service
}

func (x XService) DescribeSelf() string {
	methods := ""
	for _, m := range x.Methods() {
		methods += XMethod{m}.DescribeSelf()
	}
	return fmt.Sprintf(`
service %s {
	%s
}`,
		x.Name(),
		methods,
	)
}

type XMethod struct {
	Method
}

func (x XMethod) DescribeSelf() string {
	return fmt.Sprintf(
		`rpc %s(%s) returns (%s) {
		option %s;
	}
	`,
		x.Name(),
		x.Input().Name(),
		x.Output().Name(),
		fmtOption(x.Descriptor().GetOptions().String()),
	)
}

func fmtOption(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "[", "("), "]:", ")=")
}

type XEnum struct {
	Enum
}

func (x XEnum) DescribeSelf() string {
	s := ""
	for _, ev := range x.Values() {
		s += XEnumValue{ev}.DescribeSelf()
	}
	return fmt.Sprintf(`
enum %s {
	%s
}`,
		x.Name(),
		s)
}

type XEnumValue struct {
	EnumValue
}

func (x XEnumValue) DescribeSelf() string {
	return fmt.Sprintf("%s = %v;", x.Name(), x.Value())
}

type XMessage struct {
	Message
}

func (x XMessage) DescribeSelf() string {
	s := ""
	for _, e := range x.Enums() {
		s += XEnum{e}.DescribeSelf()
	}
	for _, m := range x.Messages() {
		s += XMessage{m}.DescribeSelf()
	}
	for _, f := range x.Fields() {
		s += XField{f}.DescribeSelf()
	}
	for _, oo := range x.OneOfs() {
		s += XOneOf{oo}.DescribeSelf()
	}
	return fmt.Sprintf(
		`message %s {
	%s
}
`,
		x.Name(),
		s)
}

type XField struct {
	Field
}

func (x XField) DescribeSelf() string {
	return fmt.Sprintf("%v %s = %v;", x.Type().ProtoType().String(), x.Name(), x.Descriptor().GetNumber())
}

type XOneOf struct {
	OneOf
}

func (x XOneOf) DescribeSelf() string {
	s := ""
	for _, f := range x.Fields() {
		s += XField{f}.DescribeSelf()
	}
	return fmt.Sprintf(`
oneof %s {
	%s
}`,
		x.Name(),
		s)
}
