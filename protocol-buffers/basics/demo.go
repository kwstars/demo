package main

import (
	"fmt"
	"log"
	"time"

	sv1 "demo/api/service/v1"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/protobuf/proto"
)

func main() {
	// 创建一个 Person 对象
	person := &sv1.Person{
		Name:  "Alice",
		Id:    1234,
		Email: "alice@example.com",
		Phones: []*sv1.Person_PhoneNumber{
			{Number: "555-4321", Type: sv1.PhoneType_PHONE_TYPE_WORK},
		},
		LastUpdated: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
		},
	}

	// 序列化 Person 对象
	out, err := proto.Marshal(person)
	if err != nil {
		log.Fatalln("Failed to encode person:", err)
	}

	// 反序列化 Person 对象
	newPerson := &sv1.Person{}
	if err := proto.Unmarshal(out, newPerson); err != nil {
		log.Fatalln("Failed to parse person:", err)
	}

	fmt.Println(newPerson)
}
