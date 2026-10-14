package disktools

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

/**

Storage device
│
├── Logical blocks (LBAs)
│   │
│   ├── Partition table (GPT)
│   │   │
│   │   └── Partition entries
│   │       │
│   │       └── Partition LBA ranges
│   │
│   └── Partitions
│       │
│       └── Filesystem or container
│           │
│           ├── Filesystem metadata
│           │   ├── Superblock/header
│           │   ├── Allocation metadata
│           │   └── Directory metadata
│           │
│           └── Directories
│               │
│               └── Files
│                   │
│                   └── File data

An LBA (Logical Block Address) is the numbered position of a fixed-size block on a storage device.
*/

type GPTHeaderExt struct {
	RawHeader  GPTHeader
	SectorSize int
}

// GPT header read directly from the disk
type GPTHeader struct {
	Signature                [8]byte
	Revision                 uint32
	HeaderSize               uint32
	HeaderCRC32              uint32
	Reserved_a               uint32
	MyLBA                    uint64
	AlternativeLBA           uint64
	FirstUsableLBA           uint64
	LastUsableLBA            uint64
	DiskGuid                 [16]byte
	PartitionEntryLBA        uint64
	NumPartitionsEntries     uint32
	SizeofPartitionEntry     uint32
	PartitionEntryArrayCRC32 uint32
}

type GPTPartitionEntry struct {
	// Unique ID that defines the purpose and type of this Partition. A value of zero defines that this partition entry is not being used.
	PartitionTypeGuid PartitionTypeGuid
	// ID unique to the partition
	UniquePartitionGuid [16]byte
	// Address of the first logical block of the partition
	StartingLBA uint64
	// address of the last logical block of the partition
	EndingLBA uint64
	// UEFI reserved bits.
	Attributes uint64
	// Human readable name of the partition
	PartitionName [72]byte
}

type PartitionTypeGuid [16]byte

func DiskIsGPT(fd int) bool {

	logicalBlockSize, err := GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("%s\n", err)
		return false
	}

	gptBytes := make([]byte, logicalBlockSize)
	rsize, err := unix.Pread(fd, gptBytes, int64(logicalBlockSize))

	if rsize <= logicalBlockSize {
		return false
	}

	return isGPTHeader(gptBytes)
}

func GetGPTHeader(fd int) (GPTHeaderExt, int) {
	gptHeaderExt := GPTHeaderExt{}
	gptHeader := GPTHeader{}
	logicalBlockSize, err := GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("Error calling ioctl to get logical block size %v\n", err)
		return gptHeaderExt, 1
	}

	gptHeaderExt.SectorSize = logicalBlockSize

	// the disk header is offset by one block
	gpt_header_bytes := make([]byte, logicalBlockSize)
	rsize, err := unix.Pread(fd, gpt_header_bytes, int64(logicalBlockSize))

	switch true {
	case err != nil:
		fmt.Println("error during pread ", err)
		return gptHeaderExt, 1
	case rsize == 0:
		fmt.Printf("Pread EOF\n")
		return gptHeaderExt, 1
	case !isGPTHeader(gpt_header_bytes):
		fmt.Printf("drive is not GPT\n")
		return gptHeaderExt, 1
	}

	ReadGPTHeader(gpt_header_bytes, &gptHeader)
	gptHeaderExt.RawHeader = gptHeader

	return gptHeaderExt, 0
}

func isGPTHeader(bytes []byte) bool {
	return string(bytes[:8]) == "EFI PART"
}

func ReadGPTEntries(fd int, GPTHeader GPTHeader, logical_block_size uint32) []GPTPartitionEntry {
	// LBA0 (protective MBR) and LBA1 (GPT header) are reserved, so start at 2
	GPTEntriesStartOffset := uint32(2)
	GPTPartitionEntries := make([]GPTPartitionEntry, 0, GPTHeader.NumPartitionsEntries)
	AllLBABytes := make([]byte, GPTHeader.SizeofPartitionEntry*GPTHeader.NumPartitionsEntries)
	_, err := unix.Pread(fd, AllLBABytes, int64(GPTEntriesStartOffset*logical_block_size))

	if err != nil {
		fmt.Printf("Error reading LBA entry: %v\n", err)
	}

	for i := uint32(0); i < GPTHeader.NumPartitionsEntries; i++ {
		LBAEntryStart := i * GPTHeader.SizeofPartitionEntry
		LBAEntryEnd := LBAEntryStart + GPTHeader.SizeofPartitionEntry
		LBABytes := AllLBABytes[LBAEntryStart:LBAEntryEnd]

		GPTPartitionEntries = append(GPTPartitionEntries, ReadLBAEntry(LBABytes))
	}

	return GPTPartitionEntries
}

func ReadLBAEntry(entryBytes []byte) GPTPartitionEntry {
	var entry GPTPartitionEntry

	if err := binary.Read(
		bytes.NewReader(entryBytes),
		binary.LittleEndian,
		&entry,
	); err != nil {
		fmt.Println(err)
	}

	return entry
}

// https://uefi.org/specs/UEFI/2.10/05_GUID_Partition_Table_Format.html#gpt-header
func ReadGPTHeader(gptBytes []byte, headerStruct *GPTHeader) {
	err := binary.Read(
		bytes.NewReader(gptBytes),
		binary.LittleEndian,
		headerStruct,
	)
	if err != nil {
		fmt.Println(err)
	}
}

func getEntryTypeString(uniquePartitionGuid PartitionTypeGuid) string {

	switch uniquePartitionGuid {
	case APFSIdentifier:
		return "APFS"
	case APFSRecoveryIdentifier:
		return "APFS Recovery"
	case APFSISCIdentifier:
		return "APFS ISC"
	}

	return "Unknown drive type"
}

func DumpGPTPartitionEntry(entry GPTPartitionEntry, sectorSize int, verbose bool) {
	if entry.StartingLBA == 0 {
		return
	}

	partitionSizeGB := ((entry.EndingLBA - entry.StartingLBA) * uint64(sectorSize)) / 1e9
	printablePartitionName := string(bytes.TrimRight(entry.PartitionName[:], "\x00"))

	if printablePartitionName == "" {
		printablePartitionName = "Untitled partition"
	}

	if !verbose {
		fmt.Printf("name: %s size: %d gb type: %s\n", printablePartitionName, partitionSizeGB, getEntryTypeString(entry.PartitionTypeGuid))
		return
	}

	fmt.Printf("----------- Partition entry -----------\n")
	fmt.Printf("PartitionSize		%d gb\n", partitionSizeGB)
	fmt.Printf("PartitionTypeGuid:       % x\n", entry.PartitionTypeGuid)
	fmt.Printf("UniquePartitionTypeGuid: % x\n", entry.UniquePartitionGuid)
	fmt.Printf("StartingLBA:              %d\n", entry.StartingLBA)
	fmt.Printf("EndingLBA:                %d\n", entry.EndingLBA)
	fmt.Printf("Attributes:               %d\n", entry.Attributes)
	fmt.Printf("PartitionName:            %s\n", printablePartitionName)

}

func DumpGPTHeader(header GPTHeaderExt, verbose bool) {

	diskSize := ((header.RawHeader.LastUsableLBA + 1) * uint64(header.SectorSize)) / 1e9

	if !verbose {
		fmt.Printf("%s	%d gb\n", header.RawHeader.Signature, diskSize)
		return
	}

	fmt.Printf("size:		     %d\n", diskSize)
	fmt.Printf("HeaderSig:           %s\n", header.RawHeader.Signature)
	fmt.Printf("Revision:           % x\n", header.RawHeader.Revision)
	fmt.Printf("HeaderSize:          %d\n", header.RawHeader.HeaderSize)
	fmt.Printf("HeaderCRC32:        % x\n", header.RawHeader.HeaderCRC32)
	fmt.Printf("Reserved_a:         % x\n", header.RawHeader.Reserved_a)
	fmt.Printf("MyLBA:              % x\n", header.RawHeader.MyLBA)
	fmt.Printf("AlternativeLBA:     % x\n", header.RawHeader.AlternativeLBA)
	fmt.Printf("FirstUsableLBA:     % x\n", header.RawHeader.FirstUsableLBA)
	fmt.Printf("LastUsableLBA:      % x\n", header.RawHeader.LastUsableLBA)
	fmt.Printf("DiskGuid:            % x\n", header.RawHeader.DiskGuid)
	fmt.Printf("PartitionEntryLBA:  % x\n", header.RawHeader.PartitionEntryLBA)
	fmt.Printf("NumPartitionEntries: %d\n", header.RawHeader.NumPartitionsEntries)
	fmt.Printf("SizeofPartitionEntr: %d\n", header.RawHeader.SizeofPartitionEntry)
	fmt.Printf("PartitionEntryCRC32: %d\n", header.RawHeader.PartitionEntryArrayCRC32)
}
