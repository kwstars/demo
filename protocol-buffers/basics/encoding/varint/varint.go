package main

import (
	"fmt"

	sv1 "demo/api/service/v1"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

func main() {
	t1 := &sv1.Person{Id: 100000}
	b, _ := proto.Marshal(t1)
	fmt.Printf("%08b, %d, %X\n", b, b, b) // [00010000 10100000 10001101 00000110], [16 160 141 6], 10A08D06
	filedTag, wireType, _ := protowire.ConsumeField(b)
	fmt.Printf("field tag: %d, wire type: %d\n", filedTag, wireType)
	v, n := protowire.ConsumeVarint(b[1:]) // 00000110 10001101 10100000 -> 00000110 0001101 0100000
	fmt.Printf("payload: %d, n: %d\n", v, n)

	// b = []byte{0x08, 0x96, 0x01} // protobuf 编码的 150，0x96 -> 1 0010110 -> 0010110，0x01 -> 0 0000001 -> 0000001，得到 00000010010110。
	// var msg wrapperspb.Int32Value
	// if err := proto.Unmarshal(b, &msg); err != nil {
	// 	fmt.Println("Failed to parse protobuf:", err)
	// 	return
	// }
	// fmt.Println("Parsed value:", msg.GetValue())

	t2 := &sv1.Person{Id: -3}
	b, _ = proto.Marshal(t2)
	fmt.Printf("%08b, %d, %X\n", b, b, b)
	fmt.Printf("field tag: %d\n", b[0]>>3)
	fmt.Printf("wire type: %d\n", b[0]&0x7)
	// 解码 varint
	varintVal, n := protowire.ConsumeVarint(b[1:])
	fmt.Printf("payload (varint): %d, n: %d\n", varintVal, n)     // payload (varint): 18446744073709551613, n: 10
	fmt.Printf("payload (signed varint): %d\n", int64(varintVal)) // payload (signed varint): -3

	fmt.Println(protowire.EncodeZigZag(-3))
}
