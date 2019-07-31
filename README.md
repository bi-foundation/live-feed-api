
# generate swagger from source

Generating a new swagger json is done with go-swagger. Go-swagger needs to be installed in $GOPATH/bin. When using go generate, the program is called to generate the spec from the comments in the code. Further information: https://goswagger.io/use/spec.html.  
```
go get -u github.com/go-swagger/go-swagger/cmd/swagger
go generate
```

