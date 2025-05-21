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
	Weight int      `json:"weight"`
	X      float64  `json:"x"`
	Y      float64  `json:"y"`
}

func PagesToJSON(pages map[string]*Page) ([]byte, error) {

	jsonPages := make(map[string]PageJSON)
	for _, page := range pages {
		jsonPages[page.link] = PageJSON{
			Title:  page.title,
			Link:   page.link,
			In:     getPageLinks(page.in),
			Out:    getPageLinks(page.out),
			Weight: page.weight,
			X:      page.x,
			Y:      page.y,
		}
	}
	return json.MarshalIndent(jsonPages, "", "  ")
}

func getPageLinks(pages []*Page) []string {
	links := make([]string, len(pages))
	for i, page := range pages {
		links[i] = page.link
	}
	return links
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
	for link, p := range rawPages {
		pages[link] = &Page{
			title:  p.Title,
			link:   p.Link,
			weight: p.Weight,
			x:      p.X,
			y:      p.Y,
			in:     []*Page{},
			out:    []*Page{},
		}
	}

	// Step 3: Second pass - assign in/out link references
	for link, p := range rawPages {
		current := pages[link]
		for _, outLink := range p.Out {
			if target, ok := pages[outLink]; ok {
				current.out = append(current.out, target)
			}
		}
		for _, inLink := range p.In {
			if source, ok := pages[inLink]; ok {
				current.in = append(current.in, source)
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
		for _, out := range page.out {
			fmt.Fprintf(f, "\"%s\" -- \"%s\";\n", page.link, out.link)
		}
	}

	fmt.Fprintln(f, "}")
	return nil
}
