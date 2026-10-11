package cli

import (
	"disk-tool/disktools"
	"flag"
	"fmt"
)

func ParseFlags() {
	listDiskFlag := flag.Bool("ld", false, "list disks")
	inspectFlag := flag.String("i", "", "inspect a drive")
	partitionsFlag := flag.String("p", "", "inspect a drive's partitions")
	verboseFlag := flag.Bool("v", false, "verbose")
	flag.Parse()

	if *inspectFlag != "" {
		isMBR, isGPT := false, false

		if disktools.DiskIsMBR(*inspectFlag) {
			isMBR = true
		}

		if disktools.DiskIsGPT(*inspectFlag) {
			isGPT = true
		}

		if isMBR && !isGPT {
			disktools.DumpMBR(disktools.ParseMBR(*inspectFlag), *verboseFlag)
			return
		}

		gptHeaderExt, errorCode := disktools.GetGPTHeader(*inspectFlag)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", *inspectFlag)
		}
		disktools.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		return
	}

	if *partitionsFlag != "" {
		gptHeaderExt, errorCode := disktools.GetGPTHeader(*partitionsFlag)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", *partitionsFlag)
		}
		disktools.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		gptEntries := disktools.ReadGPTEntries(*partitionsFlag, gptHeaderExt.RawHeader)
		for _, gptEntry := range gptEntries {
			disktools.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.SectorSize, *verboseFlag)
		}
	}

	if *listDiskFlag {
		disktools.ListDisks()
	}
}
