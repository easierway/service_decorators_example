namespace go service_decorators_example
struct Request{
  1: i32 op1,
  2: i32 op2,
}

service calculator {
  i32 add(1: Request req)
}
