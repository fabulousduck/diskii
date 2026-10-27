package disktools

import (
	diskiiOS "diskii/src/os"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func ListDisks() {
	diskNames := GetDiskNames()
	for _, diskName := range diskNames {
		fmt.Printf("%s\n", diskName)
	}
}

func GetDiskNames() []string {
	disks := []string{}
	dirEntries, err := os.ReadDir("/dev/")

	if err != nil {
		fmt.Printf("Error reading /dev/ dir %v\n", err)
		return []string{}
	}

	for _, dirEntry := range dirEntries {
		diskName := dirEntry.Name()
		if diskiiOS.IsDiskName(diskName) {
			disks = append(disks, fmt.Sprintf("/dev/%s", diskName))
		}
	}

	return disks
}

func GetFD(fileName string) int {
	fd, err := unix.Open(fileName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return fd
}
