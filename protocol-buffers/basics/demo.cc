#include <fstream>
#include <iostream>
#include <string>

#include "api/service/v1/addressbook.pb.h"

int main() {
  GOOGLE_PROTOBUF_VERIFY_VERSION;

  // 创建 Person 对象
  service::v1::Person person;
  person.set_name("John Doe");
  person.set_id(1234);
  person.set_email("jdoe@example.com");

  // 创建 Timestamp 对象
  google::protobuf::Timestamp* timestamp = new google::protobuf::Timestamp();
  timestamp->set_seconds(time(NULL));
  person.set_allocated_last_updated(timestamp);

  // 序列化 Person 对象
  std::string out;
  if (!person.SerializeToString(&out)) {
    std::cerr << "Failed to write person." << std::endl;
    return -1;
  }

  // 反序列化 Person 对象
  service::v1::Person newPerson;
  if (!newPerson.ParseFromString(out)) {
    std::cerr << "Failed to parse person." << std::endl;
    return -1;
  }

  std::cout << newPerson.DebugString() << std::endl;

  google::protobuf::ShutdownProtobufLibrary();

  return 0;
}