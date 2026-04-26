PROTOC = protoc
PROTO_PATH = ./api
CPP_OUT = ./api
JS_OUT = ./api
PROTO_FILES = $(shell find $(PROTO_PATH) -name "*.proto")
DEMO_SRC = demo.cc api/service/v1/addressbook.pb.cc
DEMO_OUT = demo

CPPFLAGS = -std=c++11 -I$(PROTO_PATH)
LDFLAGS = -lprotobuf -lpthread

.PHONY: init protoc build-cpp clean

init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

protoc:
	$(PROTOC) --proto_path=$(PROTO_PATH) --cpp_out=$(CPP_OUT) --go_out=paths=source_relative:./api --js_out=import_style=commonjs,binary:$(JS_OUT) $(PROTO_FILES)

run-cpp: protoc
	g++ $(CPPFLAGS) -o $(DEMO_OUT) $(DEMO_SRC) $(LDFLAGS)
	./$(DEMO_OUT)

run-go: protoc
	CGO_ENABLED=0 go run demo.go

clean:
	rm -f $(DEMO_OUT)