package disktools

import (
	"fmt"
	"os"
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
		if isDiskName(diskName) {
			disks = append(disks, fmt.Sprintf("/dev/%s", diskName))
		}
	}

	return disks
}
