package main

import (
	"encoding/binary"
	"fmt"
	"math"

	sv1 "demo/api/service/v1"

	"google.golang.org/protobuf/proto"
)

func main() {
	t1 := &sv1.Person{Height: 1.1}
	b, _ := proto.Marshal(t1)
	fmt.Printf("%08b, %d, %X\n", b, b, b)
	fmt.Printf("float field tag: %d\n", b[0]>>3)
	fmt.Printf("float wire type: %d\n", b[0]&0x7)
	fmt.Printf("float payload: %f\n", math.Float32frombits(binary.LittleEndian.Uint32(b[1:])))

	t2 := &sv1.Person{Height: -1.1}
	b, _ = proto.Marshal(t2)
	fmt.Printf("%08b, %d, %X\n", b, b, b)
	fmt.Printf("float field tag: %d\n", b[0]>>3)
	fmt.Printf("float wire type: %d\n", b[0]&0x7)
	fmt.Printf("float payload: %f\n", math.Float32frombits(binary.LittleEndian.Uint32(b[1:])))
}
