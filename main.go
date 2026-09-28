package main

import (
	"fmt"

	"disk-tool/disktools"
)

func main() {

	diskNames := disktools.GetDiskNames()
	for _, diskName := range diskNames {

		gptHeader, err := disktools.GetGPTHeader(diskName)
		if err != 0 {
			continue
		}

		disktools.DumpGPTHeader(gptHeader)

		entries := disktools.ReadGPTEntries(diskName, gptHeader)
		for _, entry := range entries {
			disktools.DumpGPTPartitionEntry(entry)
			switch entry.PartitionTypeGuid {
			case disktools.APFSIdentifier:

			}
		}
		fmt.Printf("\n")
	}
}
