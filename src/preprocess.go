package main

import (
	"fmt"
	"net/http"
	"net/url"
)

func CheckResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch page: %s", resp.Status)
	}

	if ct := resp.Header.Get("Content-Type"); ct == "" || ct[:9] != "text/html" {
		return fmt.Errorf("invalid content type: %s", resp.Header.Get("Content-Type"))
	}

	return nil
}

func PreprocessLinks(links []string, base string) ([]string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	var validLinks []string
	for _, link := range links {
		parsed, err := url.Parse(link)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %s", link)
		}
		resolved := baseURL.ResolveReference(parsed)
		validLinks = append(validLinks, resolved.String())
	}
	return validLinks, nil
}
