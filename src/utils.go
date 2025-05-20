package main

import "fmt"

func IsValidURL(url string) bool {
	if url == "" {
		return false
	}

	if len(url) < 8 { // minimum length for "http://"
		return false
	}

	return url[:7] == "http://" || url[:8] == "https://"
}

func PrintList(list []string, format string) {
	for _, item := range list {
		fmt.Printf(format, item)
	}
}

func PrintMap(m map[string]int, format string) {
	for key, value := range m {
		fmt.Printf(format, key, value)
	}
}
