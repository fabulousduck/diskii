package partition

import (
	"bytes"
	"diskii/src/fs"
	"diskii/src/fs/apfs"
	"diskii/src/os"
	"encoding/binary"
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
)

// Super type so we can pass info around a bit easier
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

var partitionTypeStringMap = map[string]string{
	apfs.APFS_IDENTIFIER:              "APFS",
	apfs.APFS_RECOVERY_IDENTIFIER:     "APFS Recovery",
	apfs.APFS_ISC_IDENTIFIER:          "APFS ISC",
	fs.MSFT_PARTITION_BASIC_DATA_GUID: "Microsoft basic data partition",
	fs.MSFT_PARTITION_SYSTEM_GUID:     "Microsoft system partition",
	fs.LINUX_PARTITION_IDENTIFIER:     "Linux FS partition",
}

func DiskIsGPT(fd int) bool {
	logicalBlockSize, err := os.GetLogicalBlockSize(fd)

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
	logicalBlockSize, err := os.GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("Error calling ioctl to get logical block size %v\n", err)
		return gptHeaderExt, 1
	}

	gptHeaderExt.SectorSize = logicalBlockSize
	GPTHeaderBytes := make([]byte, logicalBlockSize)
	// the disk header is offset by one block
	rsize, err := unix.Pread(fd, GPTHeaderBytes, int64(logicalBlockSize))

	switch true {
	case err != nil:
		fmt.Println("error during pread ", err)
		return gptHeaderExt, 1
	case rsize == 0:
		fmt.Printf("Pread EOF\n")
		return gptHeaderExt, 1
	case !isGPTHeader(GPTHeaderBytes):
		fmt.Printf("drive is not GPT\n")
		return gptHeaderExt, 1
	}

	ReadGPTHeader(GPTHeaderBytes, &gptHeader)
	gptHeaderExt.RawHeader = gptHeader
	return gptHeaderExt, 0
}

func isGPTHeader(bytes []byte) bool {
	return string(bytes[:8]) == "EFI PART"
}

func ReadGPTEntries(fd int, GPTHeader GPTHeader, logicalBlockSize uint32) []GPTPartitionEntry {
	// LBA0 (protective MBR) and LBA1 (GPT header) are reserved, so start at 2
	GPTEntriesStartOffset := uint32(2)
	GPTPartitionEntries := make([]GPTPartitionEntry, GPTHeader.NumPartitionsEntries)
	AllLBABytes := make([]byte, GPTHeader.SizeofPartitionEntry*GPTHeader.NumPartitionsEntries)
	_, err := unix.Pread(fd, AllLBABytes, int64(GPTEntriesStartOffset*logicalBlockSize))

	if err != nil {
		fmt.Printf("Error reading LBA entry: %v\n", err)
	}

	var wg sync.WaitGroup
	for i := uint32(0); i < GPTHeader.NumPartitionsEntries; i++ {
		wg.Add(1)
		go func(idx uint32) {
			defer wg.Done()
			LBAEntryStart := i * GPTHeader.SizeofPartitionEntry
			LBAEntryEnd := LBAEntryStart + GPTHeader.SizeofPartitionEntry
			GPTPartitionEntries[idx] = ReadLBAEntry(AllLBABytes[LBAEntryStart:LBAEntryEnd])
		}(i)
	}

	wg.Wait()

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

func getEntryTypeString(partitionTypeGuid PartitionTypeGuid) string {
	formattedTypeGuid := fmt.Sprintf("%08X-%04X-%04X-%X-%X",
		binary.LittleEndian.Uint32(partitionTypeGuid[0:4]),
		binary.LittleEndian.Uint16(partitionTypeGuid[4:6]),
		binary.LittleEndian.Uint16(partitionTypeGuid[6:8]),
		partitionTypeGuid[8:10],
		partitionTypeGuid[10:16],
	)

	for k, v := range partitionTypeStringMap {
		if formattedTypeGuid == k {
			return v
		}
	}

	return "unknown type"
}

func DumpGPTPartitionEntry(entry GPTPartitionEntry, sectorSize int, verbose bool, fsType string) string {
	if entry.StartingLBA == 0 {
		return ""
	}

	partitionSizeGB := ((entry.EndingLBA - entry.StartingLBA) * uint64(sectorSize)) / 1e9
	printablePartitionName := string(bytes.TrimRight(entry.PartitionName[:], "\x00"))

	if printablePartitionName == "" {
		printablePartitionName = "Untitled partition"
	}

	if !verbose {
		return fmt.Sprintf("fs: %s size: %d gb type: %s name: %s\n", fsType, partitionSizeGB, getEntryTypeString(entry.PartitionTypeGuid), printablePartitionName)
	}

	return fmt.Sprintf(`
	----------- Partition entry -----------
	PartitionSize		%d gb
	PartitionTypeGuid:       % x
	UniquePartitionTypeGuid: % x
	StartingLBA:              %d
	EndingLBA:                %d
	Attributes:               %d
	PartitionName:            %s
	`,
		partitionSizeGB,
		entry.PartitionTypeGuid,
		entry.UniquePartitionGuid,
		entry.StartingLBA,
		entry.EndingLBA,
		entry.Attributes,
		printablePartitionName)
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
