package fs

var MSFT_PARTITION_BASIC_DATA_GUID = "EBD0A0A2-B9E5-4433-87C0-68B6B72699C7"
var MSFT_PARTITION_SYSTEM_GUID = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"
var LINUX_PARTITION_IDENTIFIER = "0FC63DAF-8483-4772-8E79-3D69D8477DE4"

// Map of where each fs type stores its identifier.
// Values are offset from the start of the LB
var FSMap = map[string]int{
	"ReFS":  0x3,
	"NTFS":  0x3,
	"EXFAT": 0x3,
	"FAT32": 0x52,
	"FAT16": 0x36,
	"FAT12": 0x36,
}

func GetBasicPartitionFSType(logicalBlock []byte) string {

	for k, v := range FSMap {
		bytes := logicalBlock[v : v+4]
		if string(bytes) == k {
			return k
		}
	}
	return "unknown type"
}
