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

	fmt.Printf("EVAL %s\n", command)
	fmt.Printf("%b\n", command == "LISTDRIVES")

	switch command {
	case "LISTDRIVES":
		disktools.ListDisks()
	case "INSPECT":
		commandParts := strings.Split(command, " ")
		if len(commandParts) < 2 {
			fmt.Printf("no drive specified. usage: LISTDISK <DRIVE_NAME>")
		}
		gptHeader, errorCode := disktools.GetGPTHeader(commandParts[1])
		if errorCode != 0 {
			fmt.Printf("Failed to get GPT header for drive %s\n", commandParts[1])
		}
		disktools.DumpGPTHeader(gptHeader)
	}
}
