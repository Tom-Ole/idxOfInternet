package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

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

func SavePagesToFile(filename string, pages map[string]*Page) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	jsonData, _ := PagesToJSON(pages)

	_, err = file.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Pages saved to %s\n", filename)

}

func PagesToJSON(pages map[string]*Page) ([]byte, error) {
	jsonData, err := json.Marshal(pages)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func ReadPagesFromFile(filename string) (map[string]*Page, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var pages map[string]*Page
	err = json.NewDecoder(file).Decode(&pages)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Pages loaded from %s\n", filename)
	return pages, nil
}
