package ntfs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

/*
 * NTFS layout
 * Partition boot sector
 * 			|
 * Master File Table (MFT)
 * 			|
 * System files
 * 			|
 * File area
 *
 * cluster contains sectors. Everything is a sector
 *
 */

type NTFSInfo struct {
	Jump                   [3]byte // 0x00
	OEMID                  [8]byte // 0x03  "NTFS    "
	BytesPerSector         uint16  // 0x0B
	SectorsPerCluster      uint8   // 0x0D
	_                      [7]byte // 0x0E reserved/unused
	MediaDescriptor        uint8   // 0x15
	_                      [2]byte // 0x16
	SectorsPerTrack        uint16  // 0x18
	NumHeads               uint16  // 0x1A
	HiddenSectors          uint32  // 0x1C
	_                      [8]byte // 0x20
	TotalSectors           uint64  // 0x28
	MFTCluster             uint64  // 0x30
	MFTMirrCluster         uint64  // 0x38
	ClustersPerFileRecord  int8    // 0x40  signed
	_                      [3]byte
	ClustersPerIndexBuffer int8 // 0x44  signed
	_                      [3]byte
	VolumeSerial           uint64 // 0x48
}

func ReadNTFSInfo(fd int, offset int64) NTFSInfo {
	ntfsInfo := NTFSInfo{}
	// We only read the info we need. No need to read the boot sector.
	bootSectorSize := 0x50
	bootSectorBytes := make([]byte, bootSectorSize)
	_, err := unix.Pread(fd, bootSectorBytes, offset)

	if err != nil {
		return ntfsInfo
	}

	if err := binary.Read(
		bytes.NewReader(bootSectorBytes),
		binary.LittleEndian,
		&ntfsInfo,
	); err != nil {
		fmt.Println(err)
	}

	return ntfsInfo
}

func DumpNTFSInfo(ntfsInfo NTFSInfo) string {
	return fmt.Sprintf(`
		Sector size: 		 %d
		Sectors per cluster: %d
		Num sectors: 		 %d
		MFTIndex: 			 %d
		Volume serial:		 %016X
	`,
		ntfsInfo.BytesPerSector,
		ntfsInfo.SectorsPerCluster,
		ntfsInfo.TotalSectors,
		ntfsInfo.MFTCluster,
		ntfsInfo.VolumeSerial)
}
