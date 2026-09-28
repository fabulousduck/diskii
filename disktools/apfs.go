package disktools

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// docs https://developer.apple.com/support/downloads/Apple-File-System-Reference.pdf

var APFSIdentifier = PartitionTypeGuid{
	0xef, 0x57, 0x34, 0x7c,
	0x00, 0x00,
	0xaa, 0x11,
	0xaa, 0x11,
	0x00, 0x30, 0x65, 0x43, 0xec, 0xac,
}

const APFS_MAGIC_BYTES = "NXSB"

type APFSContainerSuperBlock struct {
	Checksum                       uint32
	ObjectId                       uint32
	TransactionId                  uint32
	ObjectType                     uint16
	ObjectSubType                  uint16
	ContainerMagicBytes            [4]byte
	APFSBlockSize                  uint16
	APFSBlockCount                 uint32
	OptionalFeatureFlags           uint32
	ReadOnlyCompatibleFeatureFlags uint32
	IncompatibleFeatureFlags       uint32
	ContinerUUID                   [16]byte
}

// Attempt to parse a partition as a APFS partition.
// diskName: /dev/ path to the disk to read from
// partitionEntry: GPT partition record
// LBASize: logical block size used on the disk
func ParseAPFSPartition(diskName string, paritionEntry GPTPartitionEntry, LBASize uint64) (APFSContainerSuperBlock, int) {
	var emptyBlock APFSContainerSuperBlock

	fd, err := unix.Open(diskName, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("ParseAPFSPartition could not open disk %s\n Reason: %v\n", diskName, err)
		return emptyBlock, 1
	}
	defer unix.Close(fd)

	superBlockBytes := make([]byte, LBASize)
	rsize, err := unix.Pread(fd, superBlockBytes, int64(paritionEntry.StartingLBA*LBASize))

	if rsize < int(LBASize) {
		fmt.Printf("APFS pread read less that LBA block size %d\n", LBASize)
	}

	// Validate magic bytes
	magicBytes := string(superBlockBytes[32:36])
	if magicBytes != APFS_MAGIC_BYTES {
		fmt.Printf("APFS magic bytes do not match NXSB. Got %s instead.", magicBytes)
		return emptyBlock, 1
	}

	return ReadAPFSSuperBlock(superBlockBytes), 0
}

func ReadAPFSSuperBlock(bytes []byte) APFSContainerSuperBlock {
	var block APFSContainerSuperBlock

	// LEFT OFF HERE READING THE SUPER BLOCK

	return block
}

func ReadAPFSHeader(lbaSize int, startingLBA int) {
}
