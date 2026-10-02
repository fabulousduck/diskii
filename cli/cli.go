package cli

import (
	"bufio"
	"disk-tool/disktools"
	"fmt"
	"os"
	"strings"
)

func REPL() {
	for true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">")
		text, _ := reader.ReadString('\n')
		evalCommand(strings.TrimSuffix(text, "\n"))
	}
}

func evalCommand(command string) {
	if command == "" {
		return
	}

	commandParts := strings.Split(command, " ")

	switch string(commandParts[0]) {
	case "ld":
		disktools.ListDisks()
	case "i":
		if len(commandParts) < 2 {
			fmt.Printf("no drive specified. usage: i <DRIVE_NAME>")
		}
		gptHeader, errorCode := disktools.GetGPTHeader(commandParts[1])
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", commandParts[1])
		}
		disktools.DumpGPTHeader(gptHeader)
	case "p":
		if len(commandParts) < 2 {
			fmt.Printf("no drive specified. usage: p <DRIVE_NAME>")
		}
		gptHeader, errorCode := disktools.GetGPTHeader(commandParts[1])
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", commandParts[1])
		}
		entries := disktools.ReadGPTEntries(commandParts[1], gptHeader)
		for _, entry := range entries {
			disktools.DumpGPTPartitionEntry(entry)
		}
	}
}
