package disktools

import (
	"bytes"
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

func DiskIsMBR(diskName string) bool {
	MBRSize := 512
	fd, err := unix.Open(diskName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer unix.Close(fd)

	logicalBlockSize, err := GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("%s\n", err)
		return false
	}

	mbr_bytes := make([]byte, logicalBlockSize)
	rsize, err := unix.Pread(fd, mbr_bytes, 0)

	if rsize < MBRSize {
		return false
	}

	if mbr_bytes[510] != 0x55 && mbr_bytes[511] != 0xAA {
		fmt.Printf("disk is not MBR - signature does not match 0x55AA. Got %04x %04x \n", mbr_bytes[510], mbr_bytes[511])
		return false
	}

	return true
}

func ParseMBR(diskName string) MBRStruct {
	mbrStruct := MBRStruct{}
	fd, err := unix.Open(diskName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Println(err)
		return mbrStruct
	}
	defer unix.Close(fd)

	logicalBlockSize, err := GetLogicalBlockSize(fd)

	if err != nil {
		fmt.Printf("%s\n", err)
		return mbrStruct
	}

	mbr_bytes := make([]byte, logicalBlockSize)
	unix.Pread(fd, mbr_bytes, 0)

	return ReadMBR(mbr_bytes)
}

func ReadMBR(mbrBytes []byte) MBRStruct {
	var mbrStruct MBRStruct

	if err := binary.Read(
		bytes.NewReader(mbrBytes),
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
