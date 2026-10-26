package disktools

import (
	"fmt"
	"regexp"

	"golang.org/x/sys/unix"
)

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

func isDiskName(name string) bool {
	matched, err := regexp.Match(`^disk[0-9]*$`, []byte(name))
	if err != nil {
		fmt.Printf("error regexing disk name %s\n", err.Error())

	}
	return matched
}
