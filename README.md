Go-Ecoball
-------

## Depends
You need install [protoc](https://github.com/google/protobuf/blob/master/src/README.md) 

Then you need install golang proto tools:
```bash
go get github.com/gogo/protobuf/protoc-gen-gofast
```

## Build
Run 'make all' in go-ecoball
```bash
$:~/go/src/github.com/ecoball/go-ecoball$ make
```
Then you will get a directory named 'build':
```bash
~/go/src/github.com/ecoball/go-ecoball$ ls build/
ecoball  ecoclient
```

## Notes
This project used CGO, so set the CGO_ENABLED="1"

##ecoclient
transfer aba  to another person
```
$ ./ecoclient transfer  --from=$ADDRESS --to=$ADDRESS --value=$AMOUNT
```

query account balance
```
$ ./ecoclient query balance --address=$ADDRESS
```

deploy contract,you will get contract address
```
$ ./ecoclient contract deploy --p=$CONTRACTFILE --n=$CONTRACTNAME --d=$DESCRIPTION --a=$AUTHOR --e=$EMAIL
success!
0x0133ac14c0633a2a5e09e7109dcb560f6f5270e1
```

invoke contract
```
$ ./ecoclient contract invoke -a=$CONTRACTADDRESS -m=$METHORD -p="$PARA1 $PARA2 $PARA3 ..."
```

ecoclient console
```
$ ./ecoclient --console $COMMAND
ecoclient: \> $COMMAND
...
```
##ecoball
run ecoball

```
$ ./ecoball --name=$WALLETFILE --password=$PASSWORD run
```

