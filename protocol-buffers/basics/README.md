# Protocol Buffers

![img](https://i0.wp.com/lab.wallarm.com/wp-content/uploads/2023/12/253-min.jpg)

## [Scalar Value Types](https://protobuf.dev/programming-guides/proto3/#scalar)

## Style Guide

**核心原则**：

- **一致性和可读性**： 遵循这些约定，使你的 `.proto` 文件及其对应的类保持一致且易于阅读。
- **尊重现有风格**： 在修改现有文件时，保持原有风格。
- **新文件采用最佳风格**： 创建新文件时，采用最新的最佳风格。

**标准文件格式**：

- **行长度**： 保持在 80 个字符以内。
- **缩进**： 使用 2 个空格。
- **字符串引号**： 优先使用双引号。

**文件结构**：

- **文件名**： 使用 `lower_snake_case.proto` 命名。
- **文件顺序**：
  1. 许可证头（如果适用）
  2. 文件概述
  3. 语法（`syntax`）
  4. 包（`package`）
  5. 导入（按字母顺序排序）
  6. 文件选项
  7. 其他内容

**命名规范**：

- **包名**： 小写，基于项目名称或文件路径。
- **消息名**： `PascalCase`（首字母大写），缩写也作为一个单词大写。
- **字段名**： `lower_snake_case`，包括 `oneof` 字段和扩展名。
- **枚举类型名**： `PascalCase`（首字母大写）。
- **枚举值名**： `CAPITALS_WITH_UNDERSCORES`。
- **服务名和 RPC 方法名**： `PascalCase`（首字母大写）。

**其他约定**：

- **重复字段**： 使用复数名称。
- **枚举值**： 以分号结尾，不使用逗号；零值枚举应具有 `UNSPECIFIED` 后缀。
- **避免使用**： `required` 字段（仅限 proto2）和 `groups`（仅限 proto2）。

## Enum Behavior

**核心概念**：

- **开放枚举（Open Enums）**： 允许解析和存储未知的枚举值，访问器会报告字段已设置并返回该未知值。
- **封闭枚举（Closed Enums）**： 未知的枚举值会被存储在消息的 unknown field set 中，访问器会报告字段未设置并返回默认值。

**行为差异**：

- **解析未知值**： 开放枚举直接解析并存储，封闭枚举存储在 unknown field set。
- **repeated 字段**： 封闭枚举的 repeated 字段在序列化时可能改变未知值的顺序。
- **map 字段**： 封闭枚举的 map 字段在值未知时，整个条目（键和值）会被存储在 unknown fields。

**历史背景**：

- 在 proto3 引入之前，所有枚举都是封闭的。
- Proto3 引入了开放枚举，以解决封闭枚举的意外行为。

**规范行为**：

- proto2 文件导入 proto2 枚举：封闭
- proto3 文件导入 proto3 枚举：开放
- proto3 文件导入 proto2 枚举：编译错误
- proto2 文件导入 proto3 枚举：开放

**已知问题**：

许多语言实现与规范不符，在处理 proto2 导入 proto3 枚举时，将其视为封闭枚举，包括 C++、Java、Kotlin、Objective-C (版本 < 22.0)。

**特殊情况**：

- Java：对开放枚举的处理有一些令人惊讶的边缘情况，`getName()` 和 `getNameValue()` 方法的行为不同。
- Go、JSPB、Ruby：将所有枚举视为开放。
- Dart：将所有枚举视为封闭。

**未来展望**：

Protocol Buffers 计划统一所有语言的枚举行为，以符合规范。

## Encoding

**核心概念**：

- **键值对**： Protocol Buffer 消息由一系列的键值对组成。在消息的二进制版本中，字段的编号被用作键，而字段的名称和声明类型只能在解码端通过引用消息类型的定义（即 `.proto` 文件）来确定。
- **记录**： 当一个消息被编码时，每个键值对被转换为一个由字段编号（field tag）、线类型（wire type）和有效负载（payload）组成的记录。
  - **field tag**： 是一个唯一的编号，用于在消息的二进制表示中标识和定位字段。在编码和解码时，字段的名称和类型通过引用 `.proto` 文件来确定，而字段的标签则直接编码在二进制消息中。
  - **wire type**： 用于标识字段值在二进制消息中的编码方式和数据类型的标记。例如，VARINT 线类型表示变长整数，LEN 线类型表示长度前缀的字符串或字节数组等。线类型在解码时用于确定如何读取字段的值，以及在遇到未知字段时如何跳过它。
  - **Payload**： 指字段的实际数据值，它在二进制消息中与字段的标签（field tag）和线类型（wire type）一起被编码和存储。线类型告诉解码器如何读取有效负载，而字段标签则用于标识和定位这个有效负载。
- **Tag-Length-Value (TLV)**： Protocol Buffers 编码方案。
  - Tag（标签）：这是字段的标签（field tag），用于在消息中标识和定位字段。
  - Length（长度）：这是字段值的长度。在 Protobuf 中，这个信息被线类型（wire type）隐含地表示。线类型告诉解码器如何读取字段的值，包括它的长度。
  - Value（值）：这是字段的实际数据值，也就是有效负载（payload）。

```protobuf
message Person {
    required string user_name        = 1;
    optional int64  favourite_number = 2;
    repeated string interests        = 3;
}
```

![img](https://martin.kleppmann.com/2012/12/protobuf_small.png)

**Wire Type**：

| ID  | Name   | Used For                                                 |
| --- | ------ | -------------------------------------------------------- |
| 0   | VARINT | int32, int64, uint32, uint64, sint32, sint64, bool, enum |
| 1   | I64    | fixed64, sfixed64, double                                |
| 2   | LEN    | string, bytes, embedded messages, packed repeated fields |
| 3   | SGROUP | group start (deprecated)                                 |
| 4   | EGROUP | group end (deprecated)                                   |
| 5   | I32    | fixed32, sfixed32, float                                 |

**Tag 编码**：在 Protocol Buffers（Protobuf）中，字段的标签（field tag）和线类型（wire type）被编码到一起，形成了一个称为 tag 的值。这个 tag 在二进制消息中用作键，用于标识和定位字段。

- 公式：`(field_tag << 3) | wire_type`
- 解码后，低 3 位表示 wire type，其余部分表示字段编号。
  - `wire_type = tag & 7`：提取出 tag 的低 3 位，即线类型。
  - `field_tag = tag >> 3`：将 tag 右移 3 位，即去掉低 3 位，得到字段编号。

**Protoscope**：是一个用于查看和解析 Protocol Buffers（Protobuf）消息的工具。在 Protoscope 中，Protobuf 消息的表示方式如下：

- **Tag**：Protoscope 会显示每个字段的标签（field tag）和线类型（wire type），格式为 `field tag:wire type`。例如，`1:VARINT` 表示字段编号为 1 的字段，其线类型为 VARINT。
- **Record**：Protoscope 会显示每个字段的标签、线类型和有效负载（payload），格式为 `field tag:wire type payload`。例如，`1:VARINT 150` 表示字段编号为 1 的字段，其线类型为 VARINT，有效负载为 150。
- **自动推断类型**：如果 tag 后面有空格，Protoscope 会根据下一个 token 猜测字段的类型。这是因为 Protoscope 无法访问 `.proto` 文件，所以它只能通过这种方式来尝试推断字段的类型。

需要注意的是，Protoscope 只能提供有限的信息，如果要获取字段的完整定义（包括名称、类型、是否必需等），还需要查看 `.proto` 文件。

Protocol Buffers 编码：

1. Protobuf 使用 Varints 编码和 Zigzag 编码来压缩数据。

   - **Varints 编码：** 使用变长的字节序列来表示整数，数值越小占用的字节数越少。每个字节的最高位（MSB）用作延续位，表示是否还有后续字节。
   - **Zigzag 编码：** 用于更有效地表示有符号整数。它将负数映射到正整数空间，使得绝对值较小的数具有较小的编码值。**当字段可能为负数时，建议使用 `sint32` 或 `sint64` 类型，Protobuf 会自动应用 Zigzag 编码后再进行 Varints 编码。**

2. Protobuf 不是完全自描述的。

   - **需要 proto 定义：** 接收端需要有对应的 `.proto` 文件定义才能正确解析数据。
   - **字段编号：** 序列化后的数据不携带字段名，只使用字段编号标识字段。因此，更改字段名不会影响数据解析，但**更改字段编号会破坏兼容性**。
   - **小字段编号：** 建议尽可能使用小的字段编号，因为字段编号也会被编码，较小的编号占用更少的空间。

3. Protobuf 是一种紧凑的消息结构。
   - **无字段间隔：** 编码后的字段之间没有间隔或分隔符。
   - **字段头：** 每个字段由字段编号和 wire type 组成。wire type 指示了字段值的类型和编码方式。
   - **自定界：** 字段头和 wire type 共同决定了字段值的长度，因此无需额外的标记来指示字段的结束。
   - **紧凑高效：** 这种编码方式使得 Protobuf 编码非常紧凑，传输效率高。

## Best Practices

**核心原则**：

- **向前兼容性**： 确保对 `.proto` 文件的更改不会破坏现有的客户端或服务器。
- **避免数据丢失**： 避免进行可能导致数据丢失的更改。
- **使用通用类型**： 使用内置的 [Well-Known Types](https://protobuf.dev/reference/protobuf/google.protobuf/) 和 Common Types。
- **清晰的组织结构**： 将广泛使用的消息类型定义在单独的文件中。
- **避免不必要的复杂性**： 避免创建包含大量字段的消息。
- **风格一致性**： 遵循生成的代码的风格指南。

**具体建议**：

- **不要重用标签号**： 即使字段不再使用，也不要重用其标签号。
- **为已删除的字段和枚举值保留标签号**： 使用 `reserved` 关键字。
- **不要更改字段类型**： 除非新类型是旧类型的超集。
- **不要添加必填字段 (proto2)**： 使用 `optional` 或 `repeated` 字段。
- **不要创建包含大量字段的消息**： 这会影响性能和编译。
- **在枚举中包含未指定的值**： 使用 `FOO_UNSPECIFIED` 作为第一个值。
- **不要使用 C/C++ 宏常量作为枚举值**： 可能导致编译错误。
- **使用 Well-Known Types 和 Common Types**： 例如 `duration`、`timestamp` 等。
- **将广泛使用的消息类型定义在单独的文件中**： 方便其他团队使用。
- **不要更改字段的默认值**： 可能导致客户端和服务器之间的版本偏差。
- **不要从 repeated 更改为 scalar**： 可能导致数据丢失。
- **遵循生成的代码的风格指南**： 确保生成的代码符合风格指南。
- **不要使用文本格式消息进行交换**： 文本格式在字段或枚举值重命名时容易出错。
- **不要依赖跨构建的序列化稳定性**： 不要将其用于构建缓存键等。
- **不要在同一个 Java 包中生成 JavaProtos 和其他代码**： 将生成的代码放在单独的包中。
- **避免使用语言关键字作为字段名**： 可能导致访问问题。
- **不要依赖跨构建的序列化稳定性**： 不要将其用于构建缓存键等。

## References

- https://lab.wallarm.com/what/what-is-protobuf/
- https://protobuf.dev/
- https://cloud.google.com/apis/design?hl=zh-cn
- https://martin.kleppmann.com/2012/12/05/schema-evolution-in-avro-protocol-buffers-thrift.html
