.PHONY: build-interpreter build-interpreter-debug build-schema clean

build-interpreter:
	go build -C ./interpreter/ -o ../bin/interpreter ./cmd/interpreter

build-interpreter-debug:
	go build -C ./interpreter/ -o ../bin/interpreter -gcflags="all=-N -l" ./cmd/interpreter

build-schema: build-interpreter
	./bin/interpreter

clean:
	rm -rf bin/
	rm -rf output/
