package main

import "fmt"

func main() {
	startUrl := "https://go.dev/blog/error-handling-and-go"
	root := GetPage(startUrl, 0)
	pages[startUrl] = root
	GetAllInPages()
	CalculatePageWeights()
	PrintPages()
	fmt.Print("==========================\n")
	fmt.Printf("Page length: %d \n", len(pages))
	fmt.Print("==========================\n")
}
