package disktools

var MSFT_PARTITION_BASIC_DATA_GUID = "EBD0A0A2-B9E5-4433-87C0-68B6B72699C7"
var MSFT_PARTITION_SYSTEM_GUID = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"

var FSMap = map[string]int{
	"ReFS":  0x3,
	"NTFS":  0x3,
	"EXFAT": 0x3,
	"FAT32": 0x52,
	"FAT16": 0x36,
	"FAT12": 0x36,
}

func getBasicPartitionFSType(logicalBlock []byte) string {
	for k, v := range FSMap {
		bytes := logicalBlock[v : v+8]
		if string(bytes[:]) == k {
			return k
		}
	}
	return "unknown type"
}
