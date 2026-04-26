package main

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	// 调试模式：直接从文件读取
	if os.Getenv("DEBUG") == "1" {
		debugMode()
		return
	}

	// 正常插件模式：从 stdin 读取
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	// 保存输入用于调试
	if os.Getenv("SAVE_INPUT") == "1" {
		os.WriteFile("/tmp/plugin-input.bin", input, 0644)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(input, req); err != nil {
		panic(err)
	}

	resp := Generate(req)

	output, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(output)
}

// 调试模式入口
func debugMode() {
	fmt.Println("=== DEBUG MODE ===")

	data, err := os.ReadFile("/tmp/plugin-input.bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取失败: %v\n", err)
		fmt.Fprintf(os.Stderr, "请先运行: make save\n")
		os.Exit(1)
	}

	req := &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, req); err != nil {
		fmt.Fprintf(os.Stderr, "解析失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("要生成的文件: %v\n", req.FileToGenerate)
	fmt.Printf("总文件数: %d\n\n", len(req.ProtoFile))

	// 在这里打断点！
	resp := Generate(req)

	// 输出生成的代码
	fmt.Println("========== 生成的代码 ==========")
	for _, file := range resp.File {
		fmt.Printf("\n--- %s ---\n", file.GetName())
		fmt.Println(file.GetContent())
	}
}

func Generate(req *pluginpb.CodeGeneratorRequest) *pluginpb.CodeGeneratorResponse {
	resp := &pluginpb.CodeGeneratorResponse{}

	for _, fileName := range req.FileToGenerate {
		var file *descriptorpb.FileDescriptorProto
		for _, f := range req.ProtoFile {
			if f.GetName() == fileName {
				file = f
				break
			}
		}

		if file == nil {
			continue
		}

		content := generateFile(file)

		resp.File = append(resp.File, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(fileName + ".generated.go"),
			Content: proto.String(content),
		})
	}

	resp.SupportedFeatures = proto.Uint64(uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL))

	return resp
}

func generateFile(file *descriptorpb.FileDescriptorProto) string {
	var code string

	pkgName := file.GetPackage()
	if pkgName == "" {
		pkgName = "main"
	}
	code += "package " + pkgName + "\n\n"

	for _, msg := range file.MessageType {
		code += generateMessage(msg, "")
	}

	return code
}

func generateMessage(msg *descriptorpb.DescriptorProto, prefix string) string {
	msgName := prefix + msg.GetName()
	code := "// Generated code for message: " + msgName + "\n"
	code += "type " + msgName + " struct {\n"

	for _, field := range msg.Field {
		fieldName := camelCase(field.GetName())
		fieldType := getGoType(field)
		code += "\t" + fieldName + " " + fieldType + " `json:\"" + field.GetName() + "\"`\n"
	}

	code += "}\n\n"

	for _, nested := range msg.NestedType {
		code += generateMessage(nested, msgName+"_")
	}

	return code
}

func getGoType(field *descriptorpb.FieldDescriptorProto) string {
	isRepeated := field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED

	var baseType string
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		baseType = "string"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		baseType = "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		baseType = "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		baseType = "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		baseType = "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		baseType = "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		baseType = "float32"
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		baseType = "float64"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		baseType = "[]byte"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		baseType = "*" + getTypeName(field.GetTypeName())
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		baseType = getTypeName(field.GetTypeName())
	default:
		baseType = "interface{}"
	}

	if isRepeated {
		return "[]" + baseType
	}
	return baseType
}

func getTypeName(fullName string) string {
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '.' {
			return fullName[i+1:]
		}
	}
	return fullName
}

func camelCase(s string) string {
	if s == "" {
		return ""
	}
	result := ""
	nextUpper := true
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			if c >= 'a' && c <= 'z' {
				result += string(c - 32)
			} else {
				result += string(c)
			}
			nextUpper = false
		} else {
			result += string(c)
		}
	}
	return result
}
