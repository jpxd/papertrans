package main

import (
	"bufio"
	"fmt"
	"os"

	"strings"

	"strconv"

	"golang.org/x/crypto/ssh/terminal"
)

var stdinReader *bufio.Reader

func scanInput(prompt string) string {
	if stdinReader == nil {
		stdinReader = bufio.NewReader(os.Stdin)
	}
	fmt.Print(prompt)
	bytes, _ := stdinReader.ReadString('\n')
	str := string(bytes)
	str = strings.TrimSpace(str)
	return str
}

func scanInt(prompt string) int {
	str := scanInput(prompt)
	num, _ := strconv.ParseInt(str, 10, 32)
	return int(num)
}

func scanPassword(prompt string) string {
	fmt.Print(prompt)
	pwBytes, _ := terminal.ReadPassword(0)
	fmt.Println()
	pwStr := string(pwBytes)
	pwStr = strings.TrimSpace(pwStr)
	return pwStr
}
