package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func trim_prefix(p, s string) string {
	return regexp.MustCompile("^"+p).ReplaceAllString(s, "")
}

func request_reply(message, reply string) (bool, error) {
	fmt.Printf(message, reply)

	reader := bufio.NewReader(os.Stdin)
	str, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	str = regexp.MustCompile(`[\r\n]`).ReplaceAllString(str, "")
	str = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(str, "")
	if str != "yes" {
		fmt.Println("\ngiven up.")
		return false, nil
	}

	return true, nil
}
