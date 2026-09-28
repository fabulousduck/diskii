┌───────────────────────┬────────────────────────────────┬────────────────────────────────────────────┐
│ Layer                 │ Example                        │ Communicates using                         │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 1. Application        │ Your Go GPT parser             │ Go functions and data structures           │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 2. OS API             │ unix.Open, unix.Pread          │ System calls                               │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 3. Kernel abstraction │ File descriptor for /dev/disk0 │ VFS, device nodes, major/minor IDs         │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 4. Kernel subsystem   │ macOS storage stack, IOMedia   │ Kernel-internal requests                   │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 5. Device driver      │ NVMe or USB storage driver     │ Device command protocol                    │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 6. Bus/controller     │ PCIe, USB, SATA                │ Registers, command queues, DMA, interrupts │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 7. Device firmware    │ SSD controller firmware        │ Interprets commands and controls hardware  │
├───────────────────────┼────────────────────────────────┼────────────────────────────────────────────┤
│ 8. Physical hardware  │ NAND flash cells               │ Electrical signals                         │
└───────────────────────┴────────────────────────────────┴────────────────────────────────────────────┘

For your program, a read follows approximately:

unix.Pread(fd, buffer, 4096)
        │
        ▼
System-call boundary
        │
        ▼
Kernel resolves fd → /dev/disk0
        │
        ▼
Storage subsystem creates a read request
        │
        ▼
NVMe driver submits an NVMe command
        │
        ▼
Command travels over PCIe
        │
        ▼
SSD firmware finds the requested flash data
        │
        ▼
SSD transfers data into memory using DMA
        │
        ▼
Controller signals completion with an interrupt
        │
        ▼
Kernel completes Pread
        │
        ▼
Your []byte contains the disk data

Your GPT parser operates near the top:

Physical disk bytes
→ OS block-device interface
→ []byte in your program
→ GPT header and entry decoding
→ Go structs

GPT is a data format, not a hardware protocol. The disk and driver know only that you requested bytes at an offset; your program gives those bytes meaning as GPT fields.

Hardware discovery happens through another path:

Firmware tables
    +
Bus enumeration
    +
Device descriptors
        │
        ▼
OS identifies device
        │
        ▼
OS selects matching driver
        │
        ▼
Driver exposes an OS abstraction

Examples:

Disk             → /dev/disk0
Terminal session → /dev/ttys001
Network adapter  → en0 and sockets
USB device       → I/O Kit or libusb interface
Filesystem       → directories and ordinary files

The boot sequence establishes ownership:

CPU reset vector
→ system firmware/UEFI owns basic hardware
→ firmware loads bootloader
→ bootloader loads kernel
→ OS takes ownership of devices
→ applications use OS-mediated interfaces

The most important distinction is:

Application  asks the OS
Driver       speaks the device protocol
Firmware     operates the device
Hardware     performs the physical action

Data and completion notifications then travel back upward through the same layers.


----------------------------


1. Continue reading raw sectors through  /dev/rdisk0 .
2. Use  ioctl  to query the disk’s block size and capacity.
3. Communicate with a USB device through libusb.
4. Write firmware for a small microcontroller and a host program that talks to it.
5. Experiment with a Linux kernel module inside a virtual machine.
6. Explore bare-metal programming if you want to remove the OS entirely.