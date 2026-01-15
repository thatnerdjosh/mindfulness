PREFIX ?= /usr

all: build test

build: mt

mt:
	go build -o $@ cmd/mt/main.go

test:
	go test -v ./...

run: build
	./${BINARY_NAME}

distclean: clean

clean:
	go clean
	rm -f ${BINARY_NAME}

install: mt
	@mkdir -p ${DESTDIR}${PREFIX}/bin
	# @mkdir -p ${DESTDIR}${MANPREFIX}/man1
	install mt ${DESTDIR}${PREFIX}/bin
	# install -m 644 mt.1 ${DESTDIR}${MANPREFIX}/man1

uninstall:
	rm ${DESTDIR}${PREFIX}/bin/mt
