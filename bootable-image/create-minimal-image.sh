#!/bin/bash

set -eu

VER=6.9
LINUX_SOURCE=linux-"$VER"
LINUX_TARBALL="$LINUX_SOURCE".tar.gz
LINUX_TARBALL_CHECKSUM=7beaaaf5048d1f6bcfe2006846b4dc6a8a6617426500c856fab28f9005ff36ea
LINUX_CDN=https://cdn.kernel.org/pub/linux/kernel/v6.x

install_required_packages() {
    sudo apt update
    sudo apt install build-essential ncurses-dev xz-utils libssl-dev bc flex libelf-dev bison qemu-system-x86
    sudo apt install grub-pc-bin grub-common xorriso
}

validate_kernel_tarball_sha256() {
    if [ ! -f "$LINUX_TARBALL" ]; then
        echo "no kernel tarball to validate"
        exit 1
    fi
    calculated_checksum=$(sha256sum "$LINUX_TARBALL" | awk '{print $1}')
    if [ "$calculated_checksum" != "$LINUX_TARBALL_CHECKSUM" ]; then
        echo "invalid checksum for linux tarball"
        exit 1
    fi
}

fetch_kernel_tarball() {
    if [ ! -f "$LINUX_TARBALL" ]; then
        wget "$LINUX_CDN"/"$LINUX_TARBALL" || exit 1
    fi
    validate_kernel_tarball_sha256
    if [ -e "$LINUX_SOURCE" ] && [ -d "$LINUX_SOURCE" ]; then
        rm -rf "$LINUX_SOURCE"
    fi
    echo "extracting tarball..."
    tar xf "$LINUX_TARBALL" &>/dev/null || exit 1
}

build_kernel() {
    if [ ! -e "$LINUX_SOURCE" ] || [ ! -d "$LINUX_SOURCE" ]; then
        echo "kernel source has not been fetched and unpacked"
        exit 1
    fi

    pushd "$LINUX_SOURCE" &>/dev/null || exit 1

    make distclean || exit 1
    make x86_64_defconfig || exit 1
    make kvm_guest.config || exit 1
    make -j "$(nproc)" || exit 1
    if [ ! -f arch/x86_64/boot/bzImage ]; then
        echo "kernel build failure"
        exit 1
    fi
    cp arch/x86_64/boot/bzImage /tmp/bzImage
    popd &>/dev/null || exit 1
}

build_ramdisk() {
    tempdir=$(mktemp -d)
    pushd "$tempdir" &>/dev/null || exit 1
    cat <<EOF | gcc -Wall -Wextra -pedantic --static -o init -xc - || exit 1
#include <stdio.h>
#include <unistd.h> /* sleep() */
int main()
{
	/* spin a loop in case of concurrent boot messages obscuring our message */
	while (1) {
		printf("hello world\n");
		printf("type Ctrl-a x to exit QEMU\n");
		sleep(1);
	}
}
EOF

    find . | cpio -o -H newc | gzip >/tmp/root.cpio.gz

    popd &>/dev/null || exit 1
    rm -rf "$tempdir"
}

build_iso() {
    tempdir=$(mktemp -d)
    pushd "$tempdir" &>/dev/null || exit 1

    mkdir -p iso/boot/grub
    cp -v /tmp/bzImage iso/boot/
    cp -v /tmp/root.cpio.gz iso/boot/

    cat <<EOF >iso/boot/grub/grub.cfg
set default=0
set timeout=10
menuentry 'hello world linux' --class os {
    insmod gzio
    insmod part_msdos
    linux /boot/bzImage panic=1 console=ttyS0
    initrd /boot/root.cpio.gz
}
EOF
    grub-mkrescue -o /tmp/hello-world.iso iso
    popd &>/dev/null || exit 1
}

install_required_packages
fetch_kernel_tarball
build_kernel
build_ramdisk
build_iso

qemu-system-x86_64 -nographic -no-reboot -cdrom /tmp/hello-world.iso
