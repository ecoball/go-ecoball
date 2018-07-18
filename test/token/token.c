
typedef struct {
    char * account;
    char * name;
    unsigned long long value;
} tsToken;

typedef struct {
    char *account;
    char *token;
    unsigned long long value;
} tsAccount;

int create(char * account, char * token, unsigned long long value) {
    if (RequirePermission("active") != 0) {
        return -1;
    }
    tsToken t = {0};
    t.account = account;
    t.name = token;
    t.value = value;
    if(AbaGet(token) != 0) {
        return -1;
    }
    AbaSet(token, t);
    return 0;
}

int transfer(char * from, char * to, char *token, unsigned long long value) {
    if (CheckPermission(from, "active") != 0) {
        return -1;
    }
    sub_balance(from, token, value);
    add_balance(to, token, value);
}

int sub_balance(char *account, char *token, unsigned long long value) {
    tsToken t = {0};
    tsAccount *acc = AbaGet(account);
    acc->value = acc->value - value;
    AbaSet(account, acc);
}

int add_balance(char *account, char *token, unsigned long long value) {
    tsToken t = {0};
    tsAccount *acc = AbaGet(account);
    acc->value = acc->value + value;
    AbaSet(account, acc);
}

tsAccount* balance(char *account, char *token) {
    tsToken t = {0};
    tsAccount *acc = AbaGet(account);
    return acc;
}
