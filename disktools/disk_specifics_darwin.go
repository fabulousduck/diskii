package disktools

import "golang.org/x/sys/unix"

/*
Darwin name: DKIOCGETBLOCKSIZE

source:
sdk="$(xcrun --show-sdk-path)"
less "$sdk/usr/include/sys/disk.h"
*/
func GetLogicalBlockSize(fd int) (int, error) {
	DKIOCGETBLOCKSIZE := uint(0x40046418)
	return unix.IoctlGetInt(fd, DKIOCGETBLOCKSIZE)
}
