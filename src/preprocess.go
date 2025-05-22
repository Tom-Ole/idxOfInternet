package main

import "fmt"

func PreprocessPages(pages map[string]*Page) map[string]*Page {

	fmt.Print("==========================\n")
	fmt.Printf("PreprocessPages %d\n", len(pages))
	fmt.Print("==========================\n")

	titleMap := make(map[string]*Page)  // title -> main page
	duplicates := make(map[*Page]*Page) // old -> merged into

	// Step 1: Choose representative for each title (skip empty titles)
	for _, page := range pages {
		if page.title == "" {
			continue
		}
		if _, exists := titleMap[page.title]; !exists {
			titleMap[page.title] = page
		} else {
			duplicates[page] = titleMap[page.title]
		}
	}

	// Step 2: Merge duplicate pages into the representative
	for oldPage, mainPage := range duplicates {
		// Merge in-links
		for _, inIdx := range oldPage.in {
			if inIdx >= uint32(len(pageList)) {
				continue
			}
			inPage := &pageList[inIdx]

			// Skip if already linked to main page
			alreadyLinked := false
			for _, idx := range mainPage.in {
				if idx == inIdx {
					alreadyLinked = true
					break
				}
			}
			if alreadyLinked {
				continue
			}

			// Add to main page's in-links
			mainPage.in = append(mainPage.in, inIdx)

			// Update in-page's out-links
			for i, outIdx := range inPage.out {
				if outIdx < uint32(len(pageList)) && &pageList[outIdx] == oldPage {
					// Find mainPage's index by title
					var mainIdx uint32
					for j, p := range pageList {
						if p.title == mainPage.title {
							mainIdx = uint32(j)
							break
						}
					}
					inPage.out[i] = mainIdx
				}
			}
		}

		// Merge out-links
		for _, outIdx := range oldPage.out {
			if outIdx >= uint32(len(pageList)) {
				continue
			}
			outPage := &pageList[outIdx]

			// Skip if already linked to main page
			alreadyLinked := false
			for _, idx := range mainPage.out {
				if idx == outIdx {
					alreadyLinked = true
					break
				}
			}
			if alreadyLinked {
				continue
			}

			// Add to main page's out-links
			mainPage.out = append(mainPage.out, outIdx)

			// Update out-page's in-links
			for i, inIdx := range outPage.in {
				if inIdx < uint32(len(pageList)) && &pageList[inIdx] == oldPage {
					// Find mainPage's index by title
					var mainIdx uint32
					for j, p := range pageList {
						if p.title == mainPage.title {
							mainIdx = uint32(j)
							break
						}
					}
					outPage.in[i] = mainIdx
				}
			}
		}

		// Merge weight
		mainPage.weight += oldPage.weight
	}

	// Step 3: Build new cleaned map
	cleaned := make(map[string]*Page)
	for link, page := range pages {
		if _, isDuplicate := duplicates[page]; !isDuplicate {
			cleaned[link] = page
		}
	}

	return cleaned
}
