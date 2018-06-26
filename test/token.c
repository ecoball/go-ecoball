int main() {
  AbaLogString("Create New Token");
  int ret = TokenIsExisted("Abc");
  if (ret == 1) {
    AbaLogString("The Token is Existed");
    return -1;
  }
  ret = TokenCreate("01b1a6569a557eafcccc71e0d02461fd4b601aea", "Abc", 10000);
  if (ret != 0) {
    AbaLogString("Create New Token Failed\n");
    return ret;
  }

}
