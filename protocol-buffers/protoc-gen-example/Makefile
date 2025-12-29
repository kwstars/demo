.PHONY: build save debug clean

# 编译插件
build:
	go build -o protoc-gen-example .

# 保存调试数据（只需运行一次）
save: build
	SAVE_INPUT=1 protoc --plugin=./protoc-gen-example --example_out=. testdata/test.proto
	@echo "调试数据已保存到 /tmp/plugin-input.bin"
	@echo "现在可以在 VSCode 中按 F5 调试了"

# 命令行调试
debug:
	DEBUG=1 go run .

# 清理
clean:
	rm -f protoc-gen-example /tmp/plugin-input.bin testdata/*.generated.go