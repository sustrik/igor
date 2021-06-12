.PHONY: test golden

build:
	go build -o igor/igor github.com/sustrik/igor/igor

test_tool: build
	./igor/test/test

test_lib: build
	./lib/igor/test/test

test: test_tool test_lib

golden: build
	./igor/test/update
