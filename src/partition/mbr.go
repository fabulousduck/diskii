package partition

import (
	"bytes"
	"diskii/src/os"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

type MBRStruct struct {
	BootCode               [440]byte
	UniqueMBRDiskSignature uint32
	Unknown                uint16
	PartitionRecord        [64]byte
	Signature              uint16
}

func DiskIsMBR(fd int) bool {
	MBRSize := 512
	MBRBytes := make([]byte, MBRSize)
	rsize, _ := unix.Pread(fd, MBRBytes, 0)

	if rsize < MBRSize {
		return false
	}

	return MBRBytes[510] != 0x55 && MBRBytes[511] != 0xAA
}

func ParseMBR(fd int) MBRStruct {
	mbrStruct := MBRStruct{}
	logicalBlockSize, err := os.GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("%s\n", err)
		return mbrStruct
	}

	MBRBytes := make([]byte, logicalBlockSize)
	unix.Pread(fd, MBRBytes, 0)

	if err := binary.Read(
		bytes.NewReader(MBRBytes),
		binary.LittleEndian,
		&mbrStruct,
	); err != nil {
		fmt.Println(err)
	}

	return mbrStruct
}

func DumpMBR(mbr MBRStruct, verbose bool) {
	if !verbose {
		fmt.Printf("%04X\n", mbr.UniqueMBRDiskSignature)
	}
}
