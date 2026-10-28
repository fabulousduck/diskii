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
	dumpPartitionVolumeFlag := flag.String("dv", "", "dump volume info of a partition")

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
		gptEntries := partition.ReadGPTEntries(fd, gptHeaderExt.RawHeader, uint32(gptHeaderExt.LogicalBlockSize))

		for _, gptEntry := range gptEntries {

			lbaBytes := make([]byte, gptHeaderExt.LogicalBlockSize)
			unix.Pread(fd, lbaBytes, int64(gptEntry.StartingLBA*uint64(gptHeaderExt.LogicalBlockSize)))
			possibleFSType := fs.GetBasicPartitionFSType(lbaBytes)

			if *dumpPartitionVolumeFlag != "" && (partition.FormatTypeGuid(gptEntry.UniquePartitionGuid) == *dumpPartitionVolumeFlag) {
				offset := gptEntry.StartingLBA * uint64(gptHeaderExt.LogicalBlockSize)
				fs.DumpFSInfo(fd, possibleFSType, int64(offset))
			}

			fmt.Printf("%s", partition.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.LogicalBlockSize, *verboseFlag, possibleFSType))

		}

	}

	if *listDiskFlag {
		disktools.ListDisks()
	}
}
