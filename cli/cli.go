package cli

import (
	"flag"
	"fmt"

	"golang.org/x/sys/unix"

	"disk-tool/disktools"
)

func ParseFlags() {
	listDiskFlag := flag.Bool("ld", false, "list disks")
	inspectFlag := flag.String("i", "", "inspect a drive")
	partitionsFlag := flag.String("p", "", "inspect a drive's partitions")
	verboseFlag := flag.Bool("v", false, "verbose")
	flag.Parse()

	if *inspectFlag != "" {
		isMBR, isGPT := false, false
		fd := disktools.GetFD(*inspectFlag)
		defer unix.Close(fd)

		if disktools.DiskIsMBR(fd) {
			isMBR = true
		}

		if disktools.DiskIsGPT(fd) {
			isGPT = true
		}

		if isMBR && !isGPT {
			disktools.DumpMBR(disktools.ParseMBR(fd), *verboseFlag)
			return
		}

		gptHeaderExt, errorCode := disktools.GetGPTHeader(fd)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", fd)
		}
		disktools.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		return
	}

	if *partitionsFlag != "" {
		fd := disktools.GetFD(*partitionsFlag)
		defer unix.Close(fd)
		gptHeaderExt, errorCode := disktools.GetGPTHeader(fd)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", *partitionsFlag)
		}
		disktools.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		gptEntries := disktools.ReadGPTEntries(fd, gptHeaderExt.RawHeader, uint32(gptHeaderExt.SectorSize))
		for _, gptEntry := range gptEntries {
			disktools.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.SectorSize, *verboseFlag)
		}
	}

	if *listDiskFlag {
		disktools.ListDisks()
	}
}
