package disktools

import (
	"fmt"
	"os"
	"regexp"
)

func GetDiskNames() []string {
	disks := []string{}

	dirEntries, err := os.ReadDir("/dev/")

	if err != nil {
		fmt.Printf("Error reading /dev/ dir %v\n", err)
	}

	for _, dirEntry := range dirEntries {
		diskName := dirEntry.Name()
		matched, err := regexp.Match(`^disk[0-9]+$`, []byte(diskName))
		if err != nil {
			continue
		}
		if matched {
			disks = append(disks, fmt.Sprintf("/dev/%s", diskName))
		}
	}

	return disks
}
