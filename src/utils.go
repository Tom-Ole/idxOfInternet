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

type PageJSON struct {
	Title  string   `json:"title"`
	Link   string   `json:"link"`
	In     []string `json:"in"`
	Out    []string `json:"out"`
	Weight uint16   `json:"weight"`
	X      float32  `json:"x"`
	Y      float32  `json:"y"`
	Level  uint8    `json:"level"`
}

func PagesToJSON(pages map[string]*Page) ([]byte, error) {
	jsonPages := make(map[string]PageJSON)
	for link, page := range pages {
		// Convert indices to links
		inLinks := make([]string, len(page.in))
		for i, idx := range page.in {
			if idx < uint32(len(pageList)) {
				inLinks[i] = pageList[idx].link
			}
		}

		outLinks := make([]string, len(page.out))
		for i, idx := range page.out {
			if idx < uint32(len(pageList)) {
				outLinks[i] = pageList[idx].link
			}
		}

		jsonPages[link] = PageJSON{
			Title:  page.title,
			Link:   page.link,
			In:     inLinks,
			Out:    outLinks,
			Weight: page.weight,
			X:      page.x,
			Y:      page.y,
			Level:  page.level,
		}
	}
	return json.MarshalIndent(jsonPages, "", "  ")
}

func ReadPagesFromFile(filename string) (map[string]*Page, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Step 1: Unmarshal into PageJSON map
	var rawPages map[string]PageJSON
	if err := json.Unmarshal(data, &rawPages); err != nil {
		return nil, err
	}

	// Step 2: First pass - create all Page objects without in/out links
	pages := make(map[string]*Page)
	linkToIndex := make(map[string]uint32)

	for link, p := range rawPages {
		idx := uint32(len(pageList))
		pageList = append(pageList, Page{
			title:  p.Title,
			link:   p.Link,
			weight: p.Weight,
			x:      p.X,
			y:      p.Y,
			level:  p.Level,
			in:     make([]uint32, 0),
			out:    make([]uint32, 0),
		})
		pages[link] = &pageList[idx]
		linkToIndex[link] = idx
	}

	// Step 3: Second pass - assign in/out link indices
	for link, p := range rawPages {
		current := pages[link]
		for _, outLink := range p.Out {
			if idx, ok := linkToIndex[outLink]; ok {
				current.out = append(current.out, idx)
			}
		}
		for _, inLink := range p.In {
			if idx, ok := linkToIndex[inLink]; ok {
				current.in = append(current.in, idx)
			}
		}
	}

	return pages, nil
}

func ExportToDOT(filename string, pages map[string]*Page) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "graph G {")

	for _, page := range pages {
		// Declare the node
		fmt.Fprintf(f, "\"%s\";\n", page.link)
		for _, outIdx := range page.out {
			if outIdx < uint32(len(pageList)) {
				fmt.Fprintf(f, "\"%s\" -- \"%s\";\n", page.link, pageList[outIdx].link)
			}
		}
	}

	fmt.Fprintln(f, "}")
	return nil
}
