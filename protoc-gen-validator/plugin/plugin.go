package plugin

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	valid "github.com/lanceryou/go-validator"
)

func init() {
	generator.RegisterPlugin(new(validator))
}

// spex is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for shopee spex support.
type validator struct {
	gen *generator.Generator
}

func (v *validator) Name() string {
	return "go-validator"
}

func (v *validator) Init(g *generator.Generator) {
	v.gen = g
}

func (v *validator) Generate(file *generator.FileDescriptor) {
	// only generate validator code
	v.gen.Reset()
	for _, message := range file.MessageType {
		if file.GetSyntax() == "proto3" {
			v.generateProto3Validator(file, message)
		}
	}

}

func (v *validator) GenerateImports(file *generator.FileDescriptor) {}

// P forwards to g.gen.P.
func (g *validator) P(args ...interface{}) { g.gen.P(args...) }

func (v *validator) generateProto3Validator(file *generator.FileDescriptor, desc *descriptor.DescriptorProto) {
	v.P("// generate field.")
	for _, field := range desc.Field {
		v.P(field.Name, ",type ", field.Type.String(), "type name ", field.GetTypeName(), "value ", field.GetDefaultValue(), "extend", field.GetExtendee())
	}
	v.P("// generate extend.")
	v.P()
	for _, field := range desc.Extension {
		v.P(field.Name, ",extend type ", field.Type.String(), "type name ", field.GetTypeName(), "value ", field.GetDefaultValue())
	}

	ccTypeName := generator.CamelCase(desc.GetName())
	v.P(`func (this *`, ccTypeName, `) Validate() error {`)
	// support nested message
	for _, field := range desc.Field {
		v.generateField(file, field)
	}
	v.P(`return nil`)
	v.P(`}`)
	v.P()
}

// number string 类型 按照type比较
// message 类型 生成
// reqeated 类型判断长度
func (v *validator) generateField(file *generator.FileDescriptor, field *descriptor.FieldDescriptorProto) {
	fieldValidator := getValidatorField(field)
	if fieldValidator == nil {
		return
	}

}

func getValidatorField(field *descriptor.FieldDescriptorProto) *valid.FieldValidator {
	if field.Options == nil {
		return nil
	}

	v, err := proto.GetExtension(field.Options, valid.E_Field)
	if err == nil && v.(*valid.FieldValidator) != nil {
		return v.(*valid.FieldValidator)
	}

	return nil
}
