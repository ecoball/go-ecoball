ABA
-------

## Depends
You need install protoc by 
https://github.com/google/protobuf/blob/master/src/README.md

Then you need install proto buf tools:
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