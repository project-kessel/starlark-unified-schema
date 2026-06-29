.PHONY: build-interpreter build-interpreter-debug build-schema clean test

build-interpreter:
	go build -C ./interpreter/ -o ../bin/interpreter ./cmd/interpreter

build-interpreter-debug:
	go build -C ./interpreter/ -o ../bin/interpreter -gcflags="all=-N -l" ./cmd/interpreter

test:
	go test -C ./interpreter/ -count=1 ./...

build-schema: build-interpreter
	dotenv -f .env run ./bin/interpreter

clean:
	rm -rf bin/
	rm -rf output/
