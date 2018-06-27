int create( char *addr, char * token, int max) {
  int ret = TokenIsExisted(token);
  if (1 == ret){
    return -1;
  }
  ret = TokenCreate(addr, token, max);
  if (0 != ret) {
    return -1;
  }
  return 0;
}

int transfer(char *from, char *to, char *token, long long value) {
  int ret = AbaAccountSubBalance(value, from, token);
  if (ret != 0) {
    return ret;
  }
  ret = AbaAccountAddBalance(value, from, token);
    if (ret != 0) {
    return ret;
  }
  return 0;
}

unsigned long long balance(char *addr, char *token) {
  unsigned long long value = AbaAccountGetBalance(token, addr);
  AbaLogInt(value);
  return value;
}
