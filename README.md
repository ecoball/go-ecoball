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