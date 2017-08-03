package papertrans

import (
	"bufio"
	"fmt"
	"os"

	"strings"

	"strconv"

	"github.com/howeyc/gopass"
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

func scanIntOrDefault(prompt string, defaultValue int) int {
	str := scanInput(prompt)
	if str == "" {
		return defaultValue
	}
	num, _ := strconv.ParseInt(str, 10, 32)
	return int(num)
}

func scanPassword(prompt string) string {
	pwBytes, _ := gopass.GetPasswdPrompt(prompt, true, os.Stdin, os.Stdout)
	return string(pwBytes)
}
