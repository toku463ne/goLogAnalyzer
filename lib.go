package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func argParse(args []string) map[string]string {
	m := make(map[string]string, 0)
	for _, a := range args {
		b := strings.Split(a, "=")
		if len(b) == 1 {
			m["file"] = b[0]
		} else {
			m[b[0]] = b[1]
		}
	}
	return m
}

func logmsg(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	log.Printf("%d | %s", os.Getpid(), msg)
}
