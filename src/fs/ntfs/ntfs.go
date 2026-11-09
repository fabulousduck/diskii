package ntfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	_                      [2]byte // 0x16 reserved/unused
	SectorsPerTrack        uint16  // 0x18
	NumHeads               uint16  // 0x1A
	HiddenSectors          uint32  // 0x1C
	_                      [8]byte // 0x20 reserved/unused
	TotalSectors           uint64  // 0x28
	MFTCluster             uint64  // 0x30
	MFTMirrCluster         uint64  // 0x38
	ClustersPerFileRecord  int8    // 0x40  signed
	_                      [3]byte
	ClustersPerIndexBuffer int8 // 0x44  signed
	_                      [3]byte
	VolumeSerial           uint64 // 0x48
}

type NTFS_RECORD_HEADER struct {
	Signature        [4]byte /* 0x00 "FILE" ("BAAD" = torn write)  */
	USAOffset        uint16  /* 0x04 offset of update seq. array   */
	USACount         uint16  /* 0x06 entries in USA (1 + sectors)  */
	LSN              uint64  /* 0x08 $LogFile sequence number      */
	SequenceNumber   uint16  /* 0x10 bumped each time record reused*/
	LinkCount        uint16  /* 0x12 number of hard links          */
	AttrsOffset      uint16  /* 0x14 offset of first attribute     */
	Flags            uint16  /* 0x16 MFT_RECORD_* flags            */
	BytesInUse       uint32  /* 0x18 used size of this record      */
	BytesAllocated   uint32  /* 0x1C allocated size (e.g. 1024)    */
	BaseMFTRecord    uint64  /* 0x20 0 = base record, else owner   */
	NextAttrInstance uint16  /* 0x28 next attribute ID to assign   */
	Reserved         uint16  /* 0x2A alignment (NTFS 3.1+)         */
	MFTRecordNumber  uint32  /* 0x2C this record's number (3.1+)   */
}

func ReadNTFSInfo(reader io.ReaderAt, offset int64) NTFSInfo {
	ntfsInfo := NTFSInfo{}
	// We only read the info we need. No need to read the boot code.
	bootSectorSize := 0x50
	bootSectorBytes := make([]byte, bootSectorSize)
	_, err := reader.ReadAt(bootSectorBytes, offset)

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

	if ntfsInfo.SectorsPerCluster > 0x80 {
		// have to do some type casting here because Go does not support type promotion
		// and I CBA to do the bit hacking.
		ntfsInfo.SectorsPerCluster = uint8(2 ^ (256 - int(ntfsInfo.SectorsPerCluster)))
	}

	ReadMFTFile(reader, ntfsInfo, offset)

	return ntfsInfo
}

func calcRecordSize(ntfsInfo NTFSInfo) int {
	// All calculations have to happen in Int64 due to Golangs wrapping of smaller ints
	clusterSizeBytes := int64(ntfsInfo.BytesPerSector) * int64(ntfsInfo.SectorsPerCluster)
	return int(clusterSizeBytes)
}

// At record0, we find the file header for a file called $MFT
// The contents of this file are just more file header entries.
func ReadMFTFile(reader io.ReaderAt, ntfsInfo NTFSInfo, volumeOffset int64) {
	recordZeroRelativeLocation := (int64(ntfsInfo.BytesPerSector) * int64(ntfsInfo.SectorsPerCluster)) * int64(ntfsInfo.MFTCluster)
	recordZeroAbsoluteLocation := volumeOffset + int64(recordZeroRelativeLocation)
	recordZeroStruct := NTFS_RECORD_HEADER{}
	recordZeroBytes := make([]byte, 0x48)
	_, err := reader.ReadAt(recordZeroBytes, int64(recordZeroAbsoluteLocation))

	if err != nil {
		return
	}

	if err := binary.Read(
		bytes.NewReader(recordZeroBytes),
		binary.LittleEndian,
		&recordZeroStruct,
	); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("record zero sig: %s\n", string(recordZeroStruct.Signature[:]))
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
