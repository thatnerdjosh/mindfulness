.PHONY: install podman-shell

SELINUX_ENABLED := $(shell if [ -r /sys/fs/selinux/enforce ] && [ "$$(cat /sys/fs/selinux/enforce)" = "1" ]; then echo yes; fi)
SELINUX_MOUNT_SUFFIX := $(if $(SELINUX_ENABLED),:z,)

install:
	go install ./cmd/mt

podman-shell: install
	podman run --rm -ti -v $$HOME/go/bin:/usr/local/bin$(SELINUX_MOUNT_SUFFIX) debian bash
