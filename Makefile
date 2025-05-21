all: build generate test

build:
	./build.sh

generate:
	./build_bindings.sh

test:
	./test_bindings.sh
