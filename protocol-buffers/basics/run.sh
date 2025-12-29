#!/bin/bash

set -euxo pipefail

# Change to the directory of the script
cd "$(dirname "${BASH_SOURCE[0]}")" && pwd

# Install the latest version of the protoc-gen-go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Read all .proto files into an array
readarray -d '' files < <(find api -name "*.proto" -print0)

# Use protoc to generate Go, C++, JavaScript code from the .proto files
mkdir -p ./api/cpp_out
protoc \
  --proto_path=./api \
  --go_out=paths=source_relative:./api \
  --cpp_out=./api/cpp_out \
  --js_out=import_style=commonjs,binary:./api \
  "${files[@]}"
