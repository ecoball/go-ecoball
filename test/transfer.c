int main() {
  unsigned long long balance = AbaGetAccountBalance("01b1a6569a557eafcccc71e0d02461fd4b601aea");
  AbaLogInt(balance);
  int ret = AbaAddAccountBalance(100, "01b1a6569a557eafcccc71e0d02461fd4b601aea");
  if (ret == -1) {
    AbaLogString("Failed");
  }
  AbaLogString("Success");
}
