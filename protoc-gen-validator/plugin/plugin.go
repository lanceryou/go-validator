package plugin

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	valid "github.com/lanceryou/go-validator"
	"strings"
)

func init() {
	generator.RegisterPlugin(new(validator))
}

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
	validatorMessages := getValidatorMessage(file.MessageType)
	for _, message := range file.MessageType {
		if file.GetSyntax() == "proto3" {
			v.generateProto3Validator(file, message, "", validatorMessages)
		}
	}

}

func (v *validator) GenerateImports(file *generator.FileDescriptor) {}

// P forwards to g.gen.P.
func (g *validator) P(args ...interface{}) { g.gen.P(args...) }

func (v *validator) generateProto3Validator(file *generator.FileDescriptor, desc *descriptor.DescriptorProto, prefix string, msgs []string) {
	if !hasValidatorField(desc) {
		return
	}

	ccTypeName := prefix + generator.CamelCase(desc.GetName())
	v.P(`func (this *`, ccTypeName, `) Validate() error {`)
	v.gen.In()
	// support nested message
	for _, field := range desc.Field {
		v.generateField(field, desc, msgs)
	}
	v.P(`return nil`)
	v.gen.Out()
	v.P(`}`)
	v.P()

	for _, nested := range desc.NestedType {
		v.generateProto3Validator(file, nested, ccTypeName+"_", msgs)
	}
}

// number string 类型 按照type比较
// message 类型 生成
// reqeated 类型判断长度
func (v *validator) generateField(field *descriptor.FieldDescriptorProto, desc *descriptor.DescriptorProto, msgs []string) {
	fieldValidator := getValidatorField(field)
	if fieldValidator == nil {
		return
	}

	if field.Type == nil {
		return
	}

	variableName := "this." + generator.CamelCase(*field.Name)
	if isMessage(field) && !isRepeated(field) {
		// validate nil and validate message if not nil
		v.generateMessageValidator(variableName, fieldValidator, contains(desc.GetName(), msgs))
	} else if isRepeated(field) || isString(field) {
		// validate length
		v.generateArrayValidator(variableName, fieldValidator)
	} else {
		v.generateFieldValidator(variableName, fieldValidator)
	}
}

func (v *validator) generateMessageValidator(variableName string, fv *valid.FieldValidator, isValidatorMessage bool) {
	if fv.Neq == "nil" {
		v.P(`if !(`, variableName, `!=`, fv.Neq, `) {`)
		v.gen.In()
		v.P(`return fmt.Errorf("validation error: `, variableName, ` must be not equal nil")`)
		v.gen.Out()
		v.P(`}`)
		v.P()
	}
	// if err := variableName.Validate(); err != nil{
	// 		return err
	// }
	if !isValidatorMessage {
		return
	}
	v.P(`if err := `, variableName, `.Validate(); err != nil{`)
	v.gen.In()
	v.P(`return err`)
	v.gen.Out()
	v.P(`}`)
	v.P()
}

func (v *validator) generateArrayValidator(variableName string, fv *valid.FieldValidator) {
	type Filed struct {
		Opt   string
		Value string
		Err   string
	}

	fields := []Filed{
		{
			Opt:   " < ",
			Value: fv.Lt,
			Err:   fmt.Sprintf("%s be greater than len(%s)", variableName, fv.Lt),
		},
		{
			Opt:   " > ",
			Value: fv.Gt,
			Err:   fmt.Sprintf("%s be less than len(%s)", variableName, fv.Gt),
		},
		{
			Opt:   " == ",
			Value: fv.Eq,
			Err:   fmt.Sprintf("%s be not equal len(%s)", variableName, fv.Eq),
		},
		{
			Opt:   " != ",
			Value: fv.Neq,
			Err:   fmt.Sprintf("%s be equal len(%s)", variableName, fv.Neq),
		},
	}

	for _, field := range fields {
		if field.Value != "" {
			v.P(`if !(len( `, variableName, `)`, field.Opt, field.Value, `) {`)
			v.gen.In()
			v.P(`return fmt.Errorf("validation error: `, field.Err, `")`)
			v.gen.Out()
			v.P(`}`)
			v.P()
		}
	}
}

func (v *validator) generateFieldValidator(variableName string, fv *valid.FieldValidator) {
	type Filed struct {
		Opt   string
		Value string
		Err   string
	}

	fields := []Filed{
		{
			Opt:   variableName + " < " + fv.Lt,
			Value: fv.Lt,
			Err:   fmt.Sprintf("%s be greater than %s", variableName, fv.Lt),
		},
		{
			Opt:   variableName + " > " + fv.Gt,
			Value: fv.Gt,
			Err:   fmt.Sprintf("%s be less than %s", variableName, fv.Gt),
		},
		{
			Opt:   or(variableName, " == ", fv.Eq),
			Value: fv.Eq,
			Err:   fmt.Sprintf("%s be not equal %s", variableName, fv.Eq),
		},
		{
			Opt:   variableName + " != " + fv.Neq,
			Value: fv.Neq,
			Err:   fmt.Sprintf("%s be equal %s", variableName, fv.Gt),
		},
	}

	for _, field := range fields {
		if field.Value != "" {
			v.P(`if !(`, field.Opt, `) {`)
			v.gen.In()
			v.P(`return fmt.Errorf("validation error: `, field.Err, `")`)
			v.gen.Out()
			v.P(`}`)
			v.P()
		}
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

func hasValidatorField(desc *descriptor.DescriptorProto) (has bool) {
	for _, field := range desc.Field {
		if field.Options != nil {
			return true
		}
	}

	return false
}

func getValidatorMessage(descs []*descriptor.DescriptorProto) (msgs []string) {
	for _, msg := range descs {
		if hasValidatorField(msg) {
			msgs = append(msgs, *msg.Name)
		}

		msgs = append(msgs, getValidatorMessage(msg.NestedType)...)
	}
	return
}

func or(variableName string, opt string, str string) (value string) {
	all := strings.Split(strings.Replace(str, " ", "", -1), ",")

	suffix := "||\n"
	for i, s := range all {
		if i == len(all)-1 {
			suffix = ""
		}
		value += variableName + opt + s + suffix
	}

	return value
}

func contains(dst string, src []string) bool {
	for _, str := range src {
		if str == dst {
			return true
		}
	}
	return false
}

func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

func isMessage(field *descriptor.FieldDescriptorProto) bool {
	return *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isString(field *descriptor.FieldDescriptorProto) bool {
	return *field.Type == descriptor.FieldDescriptorProto_TYPE_STRING
}
