.PHONY: build-interpreter build-interpreter-debug build-schema

build-interpreter:
	go build -C ./interpreter/ -o ../bin/interpreter cmd/interpreter/main.go

build-interpreter-debug:
	go build -C ./interpreter/ -o ../bin/interpreter -gcflags="all=-N -l" cmd/interpreter/main.go

test:
	go test -C ./interpreter/ -count=1 ./...

build-schema: build-interpreter
	./bin/interpreter