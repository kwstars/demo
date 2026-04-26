const sv1 = require("./api/service/v1/addressbook_pb.js");
const googleProtobufTimestampPb = require("google-protobuf/google/protobuf/timestamp_pb.js");

// 创建一个 Person 对象
let person = new sv1.Person();
person.setName("Alice");
person.setId(1234);
person.setEmail("alice@example.com");

let phone = new sv1.Person.PhoneNumber();
phone.setNumber("555-4321");
phone.setType(sv1.PhoneType.PHONE_TYPE_WORK);

person.setPhonesList([phone]);

// 创建一个 Timestamp 对象
let timestamp = new googleProtobufTimestampPb.Timestamp();
timestamp.setSeconds(Math.floor(Date.now() / 1000));
timestamp.setNanos(0);

person.setLastUpdated(timestamp);

// 序列化 Person 对象
let bytes = person.serializeBinary();

// 反序列化 Person 对象
let newPerson = sv1.Person.deserializeBinary(bytes);

console.log(newPerson.toObject());
