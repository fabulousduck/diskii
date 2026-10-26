package disktools

import (
	"fmt"
	"regexp"

	"golang.org/x/sys/unix"
)

var LINUX_PARTITION_IDENTIFIER = "0FC63DAF-8483-4772-8E79-3D69D8477DE4"

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
