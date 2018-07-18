int new_account(char* account, char *addr) {
    if (RequirePermission("owner") != 0) {
        return -1;
    }
    if (AbaAccountAdd(account, addr) != 0) {
        return -1;
    }
    return 0;
}

int set_account(char* account, char *perm) {
    if (RequirePermission("active") != 0) {
        return -1;
    }
    if (AddPermission(account, perm) != 0) {
        return -1;
    }
    return 0;
}

int set_contract(char* account, unsigned int type, char* description, char *code) {
    if (CheckPermission(account, "active") != 0) {
        return -1;
    }
    if (SetContract(account, type, description, code) != 0) {
        return -1;
    }
    return 0;
}

