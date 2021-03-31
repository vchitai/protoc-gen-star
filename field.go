package pgs

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

// A Field describes a member of a Message. A field may also be a member of a
// OneOf on the Message.
type Field interface {
	Entity

	// Descriptor returns the proto descriptor for this field
	Descriptor() *descriptor.FieldDescriptorProto

	// Message returns the Message containing this Field.
	Message() Message

	// InOneOf returns true if the field is in a OneOf of the parent Message.
	InOneOf() bool

	// OneOf returns the OneOf that this field is apart of. Nil is returned if
	// the field is not within a OneOf.
	OneOf() OneOf

	// Type returns the FieldType of this Field.
	Type() FieldType

	TypeName() string

	// Required returns whether or not the field is labeled as required. This
	// will only be true if the syntax is proto2.
	Required() bool

	Number() int32

	setMessage(m Message)
	setOneOf(o OneOf)
	addType(t FieldType)
}

type field struct {
	desc  *descriptor.FieldDescriptorProto
	fqn   string
	msg   Message
	oneof OneOf
	typ   FieldType

	info     SourceCodeInfo
	typeName string
	name     string
	number   int32
}

func (f *field) Name() Name {
	if len(f.name) > 0 {
		return Name(f.name)
	} else {
		return Name(f.desc.GetName())
	}
}
func (f *field) Number() int32 {
	if f.number > 0 {
		return f.number
	} else {
		return f.desc.GetNumber()
	}
}
func (f *field) FullyQualifiedName() string                   { return f.fqn }
func (f *field) Syntax() Syntax                               { return f.msg.Syntax() }
func (f *field) Package() Package                             { return f.msg.Package() }
func (f *field) Imports() []File                              { return f.typ.Imports() }
func (f *field) File() File                                   { return f.msg.File() }
func (f *field) BuildTarget() bool                            { return f.msg.BuildTarget() }
func (f *field) SourceCodeInfo() SourceCodeInfo               { return f.info }
func (f *field) Descriptor() *descriptor.FieldDescriptorProto { return f.desc }
func (f *field) Message() Message                             { return f.msg }
func (f *field) InOneOf() bool                                { return f.oneof != nil }
func (f *field) OneOf() OneOf                                 { return f.oneof }
func (f *field) Type() FieldType                              { return f.typ }
func (f *field) setMessage(m Message)                         { f.msg = m }
func (f *field) setOneOf(o OneOf)                             { f.oneof = o }

func (f *field) Required() bool {
	return f.Syntax().SupportsRequiredPrefix() &&
		f.desc.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REQUIRED
}

func (f *field) addType(t FieldType) {
	t.setField(f)
	f.typ = t
}

func (f *field) Extension(desc *proto.ExtensionDesc, ext interface{}) (ok bool, err error) {
	return extension(f.desc.GetOptions(), desc, &ext)
}

func (f *field) accept(v Visitor) (err error) {
	if v == nil {
		return
	}

	_, err = v.VisitField(f)
	return
}

func (f *field) childAtPath(path []int32) Entity {
	if len(path) == 0 {
		return f
	}
	return nil
}

func (f *field) addSourceCodeInfo(info SourceCodeInfo) { f.info = info }
func (f *field) TypeName() string {
	if len(f.typeName) > 0 {
		return f.typeName
	}
	if f.desc.TypeName != nil {
		typeName := *f.desc.TypeName
		s := strings.Split(typeName, ".")
		typeName = s[len(s)-1]
		// handle map type
		if strings.HasSuffix(typeName, "Entry") {
			var keyType, valueType string
			for _, nt := range f.msg.Descriptor().NestedType {
				if *nt.Name != typeName {
					continue
				}
				for _, fld := range nt.Field {
					if *fld.Name == "key" {
						keyType = (&field{desc: fld}).TypeName()
					}
					if *fld.Name == "value" {
						valueType = (&field{desc: fld}).TypeName()
					}
				}
			}
			if len(keyType) > 0 && len(valueType) > 0 {
				return fmt.Sprintf("map<%s,%s>", keyType, valueType)
			}
		}

		return typeName
	}

	if f.desc.Type == nil {
		return ""
	}
	return typeNameMap[*f.desc.Type]
}

var typeNameMap = map[descriptor.FieldDescriptorProto_Type]string{
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   "double",
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    "float",
	descriptor.FieldDescriptorProto_TYPE_INT64:    "int64",
	descriptor.FieldDescriptorProto_TYPE_UINT64:   "uint64",
	descriptor.FieldDescriptorProto_TYPE_INT32:    "int32",
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  "fixed64",
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  "fixed32",
	descriptor.FieldDescriptorProto_TYPE_BOOL:     "bool",
	descriptor.FieldDescriptorProto_TYPE_STRING:   "string",
	descriptor.FieldDescriptorProto_TYPE_GROUP:    "",
	descriptor.FieldDescriptorProto_TYPE_MESSAGE:  "",
	descriptor.FieldDescriptorProto_TYPE_BYTES:    "bytes",
	descriptor.FieldDescriptorProto_TYPE_UINT32:   "uint32",
	descriptor.FieldDescriptorProto_TYPE_ENUM:     "",
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: "sfixed32",
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: "sfixed64",
	descriptor.FieldDescriptorProto_TYPE_SINT32:   "sint32",
	descriptor.FieldDescriptorProto_TYPE_SINT64:   "sint64",
}
var _ Field = (*field)(nil)
