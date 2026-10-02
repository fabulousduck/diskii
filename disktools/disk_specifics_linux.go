package disktools

import (
	"fmt"
	"regexp"

	"golang.org/x/sys/unix"
)

func GetLogicalBlockSize(fd int) (int, error) {
	return unix.IoctlGetInt(fd, unix.BLKSSZGET)
}

func isDiskName(name string) bool {
	matched, err := regexp.Match(`^sd[a-z]*$`, []byte(name))
	if err != nil {
		fmt.Printf("error regexing disk name %s\n", err.Error())

	}
	return matched
}
