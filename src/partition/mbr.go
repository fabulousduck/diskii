package partition

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type MBRStruct struct {
	BootCode               [440]byte
	UniqueMBRDiskSignature uint32
	Unknown                uint16
	PartitionRecord        [64]byte
	Signature              uint16
}

func DiskIsMBR(reader io.ReaderAt) bool {
	MBRSize := 512
	MBRBytes := make([]byte, MBRSize)
	rsize, _ := reader.ReadAt(MBRBytes, 0)

	if rsize < MBRSize {
		return false
	}

	return MBRBytes[510] == 0x55 && MBRBytes[511] == 0xAA
}

// TODO: also parse the MBR mirror at the end of the disk.
func ParseMBR(reader io.ReaderAt, blockSize int) MBRStruct {
	mbrStruct := MBRStruct{}
	MBRBytes := make([]byte, blockSize)
	reader.ReadAt(MBRBytes, 0)

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
