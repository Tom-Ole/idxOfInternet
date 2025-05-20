package main

func PreprocessPages(pages map[string]*Page) map[string]*Page {
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
		for _, in := range oldPage.in {
			if in == mainPage {
				continue
			}
			mainPage.in = append(mainPage.in, in)
			// Update in.out references
			for i, out := range in.out {
				if out == oldPage {
					in.out[i] = mainPage
				}
			}
		}
		// Merge out-links
		for _, out := range oldPage.out {
			if out == mainPage {
				continue
			}
			mainPage.out = append(mainPage.out, out)
			// Update out.in references
			for i, in := range out.in {
				if in == oldPage {
					out.in[i] = mainPage
				}
			}
		}
		// Merge weight (optional)
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
