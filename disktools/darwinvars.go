package disktools

/*
source:

sdk="$(xcrun --show-sdk-path)"
less "$sdk/usr/include/sys/disk.h"
*/
const DKIOCGETBLOCKSIZE = 0x40046418
