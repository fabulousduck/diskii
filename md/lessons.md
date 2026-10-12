# Learning Progress

## Understood

- Binary fields:
  - Use fixed byte arrays for opaque, fixed-width values.
  - Decode numeric fields with the format's specified byte order.
  - Go slices use half-open ranges, and read buffers require a nonzero length.
- Disk addressing:
  - An LBA identifies a logical block, not a byte offset.
  - Convert an LBA to a whole-disk byte offset with
    `LBA * logicalBlockSize`.
  - Query the device logical-block size from the operating system rather than
    assuming it from GPT.
- macOS disk devices:
  - A `/dev/diskN` node can represent physical, virtual, or synthesized media.
  - Validate format magic before interpreting a device as GPT or APFS.
- GPT structure:
  - The primary header, partition-entry array, usable partition space, backup
    entry array, and backup header occupy distinct logical-block ranges.
  - GPT integer fields are little-endian.
  - Header fields define the partition-entry array's starting LBA, entry count,
    and on-disk entry size.
  - Read the entry array as one contiguous byte range and visit every configured
    entry; a zero type GUID marks an unused slot but does not end the array.
- GPT partition entries:
  - Entries contain type and unique GUIDs, inclusive starting and ending LBAs,
    attributes, and a UTF-16LE name.
  - Partition size is
    `(endingLBA - startingLBA + 1) * logicalBlockSize`.
  - Opening the whole disk uses whole-disk offsets, while opening a partition
    node makes offsets partition-relative.
- GPT GUIDs:
  - GPT's textual GUID representation uses mixed-endian field encoding.
  - Compare complete 16-byte GUID values in a consistent representation.
  - Byte sums or weak hashes cannot reliably identify partition types.
- A GPT partition type indicates the expected contents but does not prove that
  the partition contains a valid filesystem or container.

## Current Frontier

- Follow the discovered Apple APFS/iBoot System Container partition into its
  contents.
- Understand the hierarchy inside an APFS container before implementing it.
- Begin at partition-relative block zero and the APFS container superblock
  (`NXSB`).
- Learn the common APFS object header and the initial container-superblock
  fields, including block size, block count, feature flags, and container UUID.

## Not Yet Covered

- Validating APFS object checksums with Fletcher-64.
- Finding checkpoint descriptor areas and selecting the newest valid checkpoint.
- APFS space management and the container object map.
- Discovering and decoding APFS volume superblocks (`APSB`).
- Traversing APFS filesystem B-trees.
- Resolving directories, files, extents, extended attributes, and snapshots.
