package ntfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

var NFTS_BLOCK_SIZE = 512

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
 * cluster (NTFS concept) contains sectors (Disk concept).
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

/*
* What an MFT record looks like
*
0x000  ┌───────────────────────────┐

	│ Header: fixed fields      │  "FILE", flags, sizes, record number…

0x030  │ Header: update seq. array │  integrity check values
0x038  ├───────────────────────────┤ <- offset stored at 0x14

	│ $STANDARD_INFORMATION     │
	│ $FILE_NAME                │  Body: attributes
	│ $DATA                     │
	│ FF FF FF FF               │  end marker
	├───────────────────────────┤ <- used size (0x18)
	│ unused / slack            │

0x400  └───────────────────────────┘ <- allocated size (0x1C)
*/
type NTFSRecordHeader struct {
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

func validateUSNs(reader io.ReaderAt, ntfsRecordHeader NTFSRecordHeader, ntfsRecordOffset int64, ntfsRecordSize int) bool {
	USArrayByteCount := ntfsRecordHeader.USACount * 2
	updateSequenceArray := make([]byte, USArrayByteCount)
	MFTRecordBytes := make([]byte, ntfsRecordSize)
	reader.ReadAt(updateSequenceArray, ntfsRecordOffset+int64(ntfsRecordHeader.USAOffset))
	reader.ReadAt(MFTRecordBytes, ntfsRecordOffset)

	usn := binary.LittleEndian.Uint16(updateSequenceArray[0:2])

	// One record spans multiple blocks. Block in NTFS is always 512. Record size is usually 1024, but can differ. We do the math on it beforehand
	for i := 1; i <= (ntfsRecordSize / NFTS_BLOCK_SIZE); i++ {
		validationBytesStart := (i * 512) - 2
		validatenBytesEnd := validationBytesStart + 2
		validationBytesValue := binary.LittleEndian.Uint16(MFTRecordBytes[validationBytesStart:validatenBytesEnd])
		if validationBytesValue != usn {
			fmt.Printf("validation byte %04X does not match usn %04X", validationBytesValue, usn)
			return false
		}
	}

	return true
}

// recordSize returns the size in bytes of one MFT record.
func calcRecordSize(info NTFSInfo) int {
	clustersPerfileRecord := int8(info.ClustersPerFileRecord) // must be interpreted as signed

	if clustersPerfileRecord < 0 {
		shift := -int(clustersPerfileRecord) // e.g. 0xF6 -> -10 -> 10
		if shift < 9 || shift > 16 {         // 512 B .. 64 KB is plausible
			return 0
		}
		return 1 << shift // 2^shift bytes
	}

	clusterSize := int(info.BytesPerSector) * int(info.SectorsPerCluster)
	return int(clustersPerfileRecord) * clusterSize
}

// At record0, we find the file header for a file called $MFT
// The contents of this file are just more file header entries.
func ReadMFTFile(reader io.ReaderAt, ntfsInfo NTFSInfo, volumeOffset int64) {
	recordZeroRelativeLocation := (int64(ntfsInfo.BytesPerSector) * int64(ntfsInfo.SectorsPerCluster)) * int64(ntfsInfo.MFTCluster)
	recordZeroAbsoluteLocation := volumeOffset + int64(recordZeroRelativeLocation)
	recordZeroStruct := NTFSRecordHeader{}
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

	recordSize := calcRecordSize(ntfsInfo)
	validateUSNs(reader, recordZeroStruct, recordZeroAbsoluteLocation, recordSize)
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
