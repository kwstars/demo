package main

import (
	"fmt"

	sv1 "demo/api/service/v1"

	"google.golang.org/protobuf/proto"
)

func main() {
	t := &sv1.Person{Name: "Kira", Id: 100}
	b, _ := proto.Marshal(t)
	fmt.Printf("%08b, %d, %X\n", b, b, b)
	fmt.Printf("string field tag: %d\n", b[0]>>3)
	fmt.Printf("string wire type: %d\n", b[0]&0x7)
	fmt.Printf("string len: %d\n", b[1])
	fmt.Printf("string: %s\n", b[2:6])
	fmt.Println("int32 field tag: ", b[6]>>3)
	fmt.Println("int32 wire type: ", b[6]&0x7)
	fmt.Println("int32 payload: ", b[7])
}
