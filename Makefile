name = $(shell basename $$PWD)

build: /bin/$(name)

/bin/$(name):
	go build -o bin/$(name)

run: build
	bin/$(name)

clean:
	rm -rf bin
