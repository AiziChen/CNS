package main

import "strings"

func toAddrs(addrs string) []string {
	addrs = strings.ReplaceAll(addrs, " ", "")
	return strings.Split(addrs, ",")
}
