# go-validator


##  Generate Validate Code

    file-gen-validator 
        ast scan file
    proto-gen-validtor
        proto plugins
        

## Requirements

    Using Protobuf validators is currently verified to work with:
    
    - Go 1.11 & 1.12
    - [Protobuf](https://github.com/protocolbuffers/protobuf) @ `v3.8.0`
    - [Go Protobuf](https://github.com/golang/protobuf) @ `v1.3.2`
    
## Validator Syntax Introduce

    validator contains 4 type: lt, gt, eq, neq
    all validator field type are string.
    eq support or(like go template eq)
    
    message : 
        1: nil pointer validate, 
        2: message function validate if message set validator field.
    string or array: length validate
    number: number compare validate
 
## Example

```proto
syntax = "proto3";
package proto;
import "github.com/lanceryou/go-validator/validator.proto";

message InnerMessage {
  // some_integer can only be in range (1, 100).
  int32 some_integer = 1 [(validator.field) = {gt: "0", lt: "100"}];
  // some_float can only be in range (0;1).
  double some_float = 2 [(validator.field) = {gt: "0", lt: "1"}];
  // eq_interger can only be equal 0 or 1
  int32 eq_interger = 3 [(validator.field) = {eq: "0, 1"}];
  
  message NestMessage{
    int32 some_integer = 1 [(validator.field) = {gt: "0", lt: "100"}];
  }
  
  // support nest message
  // nest_message can not equal nil
  // nest_message.some_integer can only be in range (1, 100).
  NestMessage nest_message = 5[(validator.field) = {neq:"nil"}];
}

```  

```go
func (this *InnerMessage) Validate() error {
	if !(this.SomeInteger < 100) {
		return fmt.Errorf("validation error: this.SomeInteger be greater than 100")
	}

	if !(this.SomeInteger < 0) {
		return fmt.Errorf("validation error: this.SomeInteger be less than 0")
	}

	if !(this.SomeFloat < 1) {
		return fmt.Errorf("validation error: this.SomeFloat be greater than 1")
	}

	if !(this.SomeFloat < 0) {
		return fmt.Errorf("validation error: this.SomeFloat be less than 0")
	}

	if !(this.EqInterger == 0 ||
		this.EqInterger == 1) {
		return fmt.Errorf("validation error: this.EqInterger be not equal 0, 1")
	}

	if !(this.NestMessage != nil) {
		return fmt.Errorf("validation error: this.NestMessage must be not equal nil")
	}

	if err := this.NestMessage.Validate(); err != nil {
		return err
	}

	return nil
}

func (this *InnerMessage_NestMessage) Validate() error {
	if !(this.SomeInteger < 100) {
		return fmt.Errorf("validation error: this.SomeInteger be greater than 100")
	}

	if !(this.SomeInteger < 0) {
		return fmt.Errorf("validation error: this.SomeInteger be less than 0")
	}

	return nil
}
```

## Installing and using

 go get github.com/lanceryou/go-validator/protoc-gen-govalidator
 
  protoc --proto_path=${GOPATH}/src --proto_path=. \
    --validator_out=plugins=go-validator:. *.proto