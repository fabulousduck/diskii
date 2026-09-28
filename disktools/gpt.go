package disktools

import (
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
	Reserved_b               []byte
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
	// Reserved by UEFI. Must be 0
	Reserved []byte
}

type PartitionTypeGuid [16]byte

func GetGPTHeader(diskName string) (GPTHeader, int) {
	gptHeader := GPTHeader{}
	fd, err := unix.Open(diskName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return gptHeader, 1
	}
	defer unix.Close(fd)

	logicalBlockSize, err := unix.IoctlGetInt(fd, DKIOCGETBLOCKSIZE)

	if err != nil {
		fmt.Printf("Error calling ioctl DKIOCGETBLOCKSIZE to get logical block size %v\n", err)
		return gptHeader, 1
	}

	gpt_header_bytes := make([]byte, logicalBlockSize)

	// the disk header is offset by one block
	rsize, err := unix.Pread(fd, gpt_header_bytes, int64(logicalBlockSize))

	switch true {
	case err != nil:
		fmt.Println(err)
		return gptHeader, 1
	case rsize == 0:
		fmt.Println("Pread EOF")
		return gptHeader, 1
	case !isGPTHeader(gpt_header_bytes):
		return gptHeader, 1
	}

	ReadGPTHeader(gpt_header_bytes, &gptHeader)

	return gptHeader, 0
}

func ReadGPTEntries(diskName string, gptHeader GPTHeader) []GPTPartitionEntry {
	fd, err := unix.Open(diskName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return []GPTPartitionEntry{}
	}
	defer unix.Close(fd)
	logicalBlockSize, err := unix.IoctlGetInt(fd, DKIOCGETBLOCKSIZE)

	return readGPTEntries(fd, &gptHeader, uint32(logicalBlockSize))
}

func isGPTHeader(bytes []byte) bool {
	return string(bytes[:8]) == "EFI PART"
}

func readGPTEntries(fd int, GPTHeader *GPTHeader, logical_block_size uint32) []GPTPartitionEntry {
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

func ReadLBAEntry(bytes []byte) GPTPartitionEntry {
	var entry GPTPartitionEntry

	copy(entry.PartitionTypeGuid[:], bytes[0:16])
	copy(entry.UniquePartitionGuid[:], bytes[16:32])
	entry.StartingLBA = binary.LittleEndian.Uint64(bytes[32:40])
	entry.EndingLBA = binary.LittleEndian.Uint64(bytes[40:48])
	entry.Attributes = binary.LittleEndian.Uint64(bytes[48:56])
	copy(entry.PartitionName[:], bytes[56:128])

	return entry
}

// https://uefi.org/specs/UEFI/2.10/05_GUID_Partition_Table_Format.html#gpt-header
func ReadGPTHeader(bytes []byte, headerStruct *GPTHeader) {

	copy(headerStruct.Signature[:], bytes[0:8])
	headerStruct.Revision = binary.LittleEndian.Uint32(bytes[8:12])
	headerStruct.HeaderSize = binary.LittleEndian.Uint32(bytes[12:16])
	headerStruct.HeaderCRC32 = binary.LittleEndian.Uint32(bytes[16:20])
	headerStruct.Reserved_a = binary.LittleEndian.Uint32(bytes[20:24])
	headerStruct.MyLBA = binary.LittleEndian.Uint64(bytes[24:32])
	headerStruct.AlternativeLBA = binary.LittleEndian.Uint64(bytes[32:40])
	headerStruct.FirstUsableLBA = binary.LittleEndian.Uint64(bytes[40:48])
	headerStruct.LastUsableLBA = binary.LittleEndian.Uint64(bytes[48:56])
	copy(headerStruct.DiskGuid[:], bytes[56:72])
	headerStruct.PartitionEntryLBA = binary.LittleEndian.Uint64(bytes[72:80])
	headerStruct.NumPartitionsEntries = binary.LittleEndian.Uint32(bytes[80:84])
	headerStruct.SizeofPartitionEntry = binary.LittleEndian.Uint32(bytes[84:88])
	headerStruct.PartitionEntryArrayCRC32 = binary.LittleEndian.Uint32(bytes[88:92])
	copy(headerStruct.Reserved_b[:], bytes[92:len(bytes)-1])
}

func DumpGPTPartitionEntry(entry GPTPartitionEntry) {
	if entry.StartingLBA == 0 {
		return
	}

	fmt.Printf("----------- Partition entry -----------\n")
	fmt.Printf("PartitionTypeGuid:       % x\n", entry.PartitionTypeGuid)
	fmt.Printf("UniquePartitionTypeGuid: % x\n", entry.UniquePartitionGuid)
	fmt.Printf("StartingLBA:              %d\n", entry.StartingLBA)
	fmt.Printf("EndingLBA:                %d\n", entry.EndingLBA)
	fmt.Printf("Attributes:               %d\n", entry.Attributes)
	fmt.Printf("PartitionName:            %s\n", entry.PartitionName)
}

func DumpGPTHeader(header GPTHeader) {
	fmt.Printf("HeaderSig:           %s\n", header.Signature)
	fmt.Printf("Revision:           % x\n", header.Revision)
	fmt.Printf("HeaderSize:          %d\n", header.HeaderSize)
	fmt.Printf("HeaderCRC32:        % x\n", header.HeaderCRC32)
	fmt.Printf("Reserved_a:         % x\n", header.Reserved_a)
	fmt.Printf("MyLBA:              % x\n", header.MyLBA)
	fmt.Printf("AlternativeLBA:     % x\n", header.AlternativeLBA)
	fmt.Printf("FirstUsableLBA:     % x\n", header.FirstUsableLBA)
	fmt.Printf("LastUsableLBA:      % x\n", header.LastUsableLBA)
	fmt.Printf("DiskGuid:            % x\n", header.DiskGuid)
	fmt.Printf("PartitionEntryLBA:  % x\n", header.PartitionEntryLBA)
	fmt.Printf("NumPartitionEntries: %d\n", header.NumPartitionsEntries)
	fmt.Printf("SizeofPartitionEntr: %d\n", header.SizeofPartitionEntry)
	fmt.Printf("PartitionEntryCRC32: %d\n", header.PartitionEntryArrayCRC32)
	fmt.Printf("Reserved_b:          % x\n", header.Reserved_b)
}
