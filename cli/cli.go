package cli

import (
	disktools "diskii/src"
	"diskii/src/fs"
	"diskii/src/partition"
	"flag"
	"fmt"

	"golang.org/x/sys/unix"
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

		if partition.DiskIsMBR(fd) {
			isMBR = true
		}

		if partition.DiskIsGPT(fd) {
			isGPT = true
		}

		if isMBR && !isGPT {
			partition.DumpMBR(partition.ParseMBR(fd), *verboseFlag)
			return
		}

		gptHeaderExt, errorCode := partition.GetGPTHeader(fd)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %d\n", fd)
		}
		partition.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		return
	}

	if *partitionsFlag != "" {
		fd := disktools.GetFD(*partitionsFlag)
		defer unix.Close(fd)
		gptHeaderExt, errorCode := partition.GetGPTHeader(fd)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", *partitionsFlag)
		}
		partition.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		gptEntries := partition.ReadGPTEntries(fd, gptHeaderExt.RawHeader, uint32(gptHeaderExt.SectorSize))
		for _, gptEntry := range gptEntries {
			lbaBytes := make([]byte, gptHeaderExt.SectorSize)
			unix.Pread(fd, lbaBytes, int64(gptEntry.StartingLBA*uint64(gptHeaderExt.SectorSize)))
			possibleFSType := fs.GetBasicPartitionFSType(lbaBytes)
			partition.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.SectorSize, *verboseFlag, possibleFSType)
		}
	}

	if *listDiskFlag {
		disktools.ListDisks()
	}
}
