int new_account(char* name, char *addr) {
  if (RequirePermission("owner") != 0) {
    return -1;
  }
  if (AbaAccountAdd(name, addr) != 0) {
    return -1;
  }
  return 0;
}

int set_account(char* name, char *perm) {
  if (RequirePermission("active") != 0) {
    return -1;
  }
  if (AddPermission(name, perm) != 0) {
    return -1;
  }
  return 0;
}

int transfer(char* from, char* to, unsigned long long value) {
  if (CheckPermission(from, "active") != 0) {
    return -1;
  }
  if(AbaAccountSubBalance(from, value) != 0) {
    return -1;
  }
  if(AbaAccountAddBalance(to, value) != 0) {
    return -1;
  }
  return 0;
}

