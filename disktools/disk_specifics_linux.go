package disktools

import "golang.org/x/sys/unix"

func GetLogicalBlockSize(fd int) (int, error) {
	return unix.IoctlGetInt(fd, unix.BLKSSZGET)
}
