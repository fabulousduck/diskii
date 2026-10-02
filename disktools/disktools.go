package disktools

import (
	"fmt"
	"os"
	"regexp"
)

func ListDisks() {
	diskNames := GetDiskNames()
	for _, diskName := range diskNames {
		fmt.Printf("%s\n", diskName)
	}
}

func InspectDisk(diskName string) {
	gptHeader, err := GetGPTHeader(diskName)
	if err != 0 {
		fmt.Printf("Error reading disk %s. Error: %s\n", diskName, err)
		return
	}

	DumpGPTHeader(gptHeader)

	entries := ReadGPTEntries(diskName, gptHeader)
	for _, entry := range entries {
		DumpGPTPartitionEntry(entry)
	}
}

func GetDiskNames() []string {
	disks := []string{}

	dirEntries, err := os.ReadDir("/dev/")

	if err != nil {
		fmt.Printf("Error reading /dev/ dir %v\n", err)
	}

	for _, dirEntry := range dirEntries {
		diskName := dirEntry.Name()
		matched, err := regexp.Match(`^disk[0-9]*$`, []byte(diskName))
		if err != nil {
			fmt.Printf("error regexing disk name %s\n", err.Error())
			continue
		}
		if matched {
			disks = append(disks, fmt.Sprintf("/dev/%s", diskName))
		}
	}

	return disks
}
