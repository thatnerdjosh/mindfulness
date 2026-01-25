PREFIX ?= /usr
BINARY_NAME ?= mt

.PHONY: all build test run clean distclean install install-prefix uninstall podman-shell

SELINUX_ENABLED := $(shell if [ -r /sys/fs/selinux/enforce ] && [ "$$(cat /sys/fs/selinux/enforce)" = "1" ]; then echo yes; fi)
SELINUX_MOUNT_SUFFIX := $(if $(SELINUX_ENABLED),:z,)

all: build test

build: $(BINARY_NAME)

$(BINARY_NAME):
	go build -o $@ cmd/mt/main.go

test:
	go test -v ./...

run: build
	./$(BINARY_NAME)

clean:
	go clean
	rm -f $(BINARY_NAME)

distclean: clean

install:
	go install ./cmd/mt

install-prefix: $(BINARY_NAME)
	@mkdir -p $(DESTDIR)$(PREFIX)/bin
	install $(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)

podman-shell: install
	podman run --rm -ti -v $$HOME/go/bin:/usr/local/bin$(SELINUX_MOUNT_SUFFIX) debian bash
