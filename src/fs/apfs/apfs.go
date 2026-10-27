package apfs

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

// source: https://developer.apple.com/support/downloads/Apple-File-System-Reference.pdf

var APFS_IDENTIFIER = "7C3457EF-0000-11AA-AA11-00306543ECAC"
var APFS_RECOVERY_IDENTIFIER = "52637672-7900-11AA-AA11-00306543ECAC"
var APFS_ISC_IDENTIFIER = "69646961-6700-11AA-AA11-00306543ECAC"

const APFS_MAGIC_BYTES = "BSXN"
const MAX_CKSUM_SIZE = 8
const NX_MAX_FILE_SYSTEMS = 100
const NX_EPH_INFO_COUNT = 4
const NX_EPH_MIN_BLOCK_COUNT = 8
const NX_MAX_FILE_SYSTEM_EPH_STRUCTS = 4
const NX_TX_MIN_CHECKPOINT_COUNT = 4
const NX_EPH_INFO_VERSION_1 = 1
const NX_NUM_COUNTERS = 32

type obj_phys_t struct {
	o_cksum   [MAX_CKSUM_SIZE]uint8
	o_oid     uint64
	o_xid     uint64
	o_type    uint32
	o_subtype uint32
}

type prange_t struct {
	pr_start_paddr uint64
	pr_block_count uint64
}

type uuid_t [16]uint8

type APFSContainerSuperBlock struct {
	nx_o                            obj_phys_t
	nx_magic                        uint32
	nx_block_size                   uint32
	nx_block_count                  uint64
	nx_features                     uint64
	nx_readonly_compatible_features uint64
	nx_incompatible_features        uint64
	nx_uuid                         uuid_t
	nx_next_oid                     uint64
	nx_next_xid                     uint64
	nx_xp_desc_blocks               uint32
	nx_xp_data_blocks               uint32
	nx_xp_desc_base                 uint64
	nx_xp_data_base                 uint64
	nx_xp_desc_next                 uint32
	nx_xp_data_next                 uint32
	nx_xp_desc_index                uint32
	nx_xp_desc_len                  uint32
	nx_xp_data_index                uint32
	nx_xp_data_len                  uint32
	nx_spaceman_oid                 uint64
	nx_omap_oid                     uint64
	nx_reaper_oid                   uint64
	nx_test_type                    uint32
	nx_max_file_systems             uint32
	nx_fs_oid                       [NX_MAX_FILE_SYSTEMS]uint64
	nx_counters                     [NX_NUM_COUNTERS]uint64
	nx_blocked_out_prange           prange_t
	nx_evict_mapping_tree_oid       uint64
	nx_flags                        uint64
	nx_efi_jumpstart                uint64
	nx_fusion_uuid                  uuid_t
	nx_keylocker                    prange_t
	nx_ephemeral_info               [NX_EPH_INFO_COUNT]uint64
	nx_test_oid                     uint64
	nx_fusion_mt_oid                uint64
	nx_fusion_wbc_oid               uint64
	nx_fusion_wbc                   prange_t
	nx_newest_mounted_version       uint64
	nx_mkb_locker                   prange_t
}

// Attempt to parse a partition as a APFS partition.
func ParseAPFSPartition(fd int, startingLBA uint64, LBASize uint64) (APFSContainerSuperBlock, int) {
	var emptyBlock APFSContainerSuperBlock
	superBlockBytes := make([]byte, LBASize)
	rsize, err := unix.Pread(fd, superBlockBytes, int64(startingLBA*LBASize))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	if rsize < int(LBASize) {
		fmt.Printf("APFS pread read less that LBA block size %d\n", LBASize)
	}

	// Validate magic bytes
	magicBytes := string(superBlockBytes[32:36])
	if magicBytes != APFS_MAGIC_BYTES {
		fmt.Printf("APFS magic bytes do not match NXSB. Got %s instead.", magicBytes)
		return emptyBlock, 1
	}

	block := ReadAPFSSuperBlock(superBlockBytes)
	DumpAPFSSuperblock(block)

	return block, 0
}

func ReadAPFSSuperBlock(superblockBytes []byte) APFSContainerSuperBlock {
	var block APFSContainerSuperBlock

	if err := binary.Read(
		bytes.NewReader(superblockBytes),
		binary.LittleEndian,
		&block,
	); err != nil {
		fmt.Println(err)
	}

	return block
}

func DumpAPFSSuperblock(block APFSContainerSuperBlock) {
	fmt.Printf("magic: %04X\n blk size: %d\b blk count: %d\n", block.nx_magic, block.nx_block_size, block.nx_block_count)
}
