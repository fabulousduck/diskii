package cli

import (
	disktools "diskii/src"
	"diskii/src/fs"
	"diskii/src/os"
	"diskii/src/partition"
	"flag"
	"fmt"
	"io"

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

		reader := disktools.GetReaderAt(*inspectFlag)

		blockSize, err := os.GetLogicalBlockSize(fd)
		if err != nil {
			fmt.Printf("Error calling ioctl to get logical block size %v\n", err)
			return
		}

		if partition.DiskIsMBR(reader) {
			isMBR = true
		}

		if partition.DiskIsGPT(reader, blockSize) {
			isGPT = true
		}

		if isMBR && !isGPT {
			partition.DumpMBR(partition.ParseMBR(reader, blockSize), *verboseFlag)
			return
		}

		gptHeaderExt, errorCode := partition.GetGPTHeader(reader, blockSize)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %d\n", fd)
		}
		partition.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		return
	}

	if *partitionsFlag != "" {
		fd := disktools.GetFD(*partitionsFlag)
		defer unix.Close(fd)
		blockSize, err := os.GetLogicalBlockSize(fd)
		reader := disktools.GetReaderAt(*partitionsFlag)

		if err != nil {
			return
		}

		gptHeaderExt, errorCode := partition.GetGPTHeader(reader, blockSize)
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", *partitionsFlag)
		}
		partition.DumpGPTHeader(gptHeaderExt, *verboseFlag)
		gptEntries := partition.ReadGPTEntries(reader, gptHeaderExt.RawHeader, uint32(gptHeaderExt.LogicalBlockSize))

		for _, gptEntry := range gptEntries {

			lbaBytes := make([]byte, gptHeaderExt.LogicalBlockSize)
			unix.Pread(fd, lbaBytes, int64(gptEntry.StartingLBA*uint64(gptHeaderExt.LogicalBlockSize)))
			possibleFSType := fs.GetBasicPartitionFSType(lbaBytes)

			if *dumpPartitionVolumeFlag == "" {
				fmt.Printf("%s", partition.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.LogicalBlockSize, *verboseFlag, possibleFSType))
			}

			if *dumpPartitionVolumeFlag != "" && (partition.FormatTypeGuid(gptEntry.UniquePartitionGuid) == *dumpPartitionVolumeFlag) {
				offset := gptEntry.StartingLBA * uint64(gptHeaderExt.LogicalBlockSize)
				fs.DumpFSInfo(reader, possibleFSType, int64(offset))
				fmt.Printf("%s", partition.DumpGPTPartitionEntry(gptEntry, gptHeaderExt.LogicalBlockSize, *verboseFlag, possibleFSType))
			}

		}

	}

	if *listDiskFlag {
		disktools.ListDisks()
	}
}

func PrettyDump(reader io.ReaderAt) {
}
