package main

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"mvdan.cc/xurls/v2"
)

func TestMarkdown(t *testing.T) {
    t.Run("URLs", validateURLs)
    t.Run("Headers", validateReadmeHeaders)
    t.Run("NotEmpty", validateReadmeNotEmpty)
    t.Run("ResourceTableHeaders", validateResourceTableHeaders)
    t.Run("InputsTableHeaders", validateInputsTableHeaders)
    t.Run("OutputsTableHeaders", validateOutputsTableHeaders)
}

func validateURLs(t *testing.T) {
	readmePath := os.Getenv("README_PATH")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to load markdown file: %v", err)
	}

	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(string(data), -1)

	var wg sync.WaitGroup
	for _, u := range urls {
		// Skip Terraform Registry URLs because of status 200 on non-existing URLs.
		if strings.Contains(u, "registry.terraform.io/providers/") {
			continue
		}

		// Parse the URL before making a request
		_, err := url.Parse(u)
		if err != nil {
			continue
		}

		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			resp, err := http.Get(url)
			if err != nil {
				t.Errorf("Failed: URL: %s, Error: %v", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Failed: URL: %s, Status code: %d", url, resp.StatusCode)
			} else {
				t.Logf("Success: URL: %s, Status code: %d", url, resp.StatusCode)
			}
		}(u)
	}
	wg.Wait()
}

func validateReadmeHeaders(t *testing.T) {
	readmePath := os.Getenv("README_PATH")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to load markdown file: %v", err)
	}

	contents := string(data)

	requiredHeaders := map[string]int{
		"## Goals":     1,
		"## Resources": 1,
		"## Inputs":    1,
		"## Outputs":   1,
		"## Features":  1,
		"## Testing":   1,
		"## Authors":   1,
		"## License":   1,
		"## Usage":     2,
	}

	for header, minCount := range requiredHeaders {
		matches := regexp.MustCompile("(?m)^"+regexp.QuoteMeta(header)).FindAllString(contents, -1)
		if len(matches) < minCount {
			t.Errorf("Failed: README.md does not contain required header '%s' at least %d times", header, minCount)
		} else {
			t.Logf("Success: README.md contains required header '%s' at least %d times", header, minCount)
		}
	}
}

func validateReadmeNotEmpty(t *testing.T) {
	readmePath := os.Getenv("README_PATH")

	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed: Cannot access README.md: %v", err)
	}

	t.Log("Success: README.md file exists.")

	if len(data) == 0 {
		t.Errorf("Failed: README.md is empty.")
	} else {
		t.Log("Success: README.md is not empty.")
	}
}

func validateResourceTableHeaders(t *testing.T) {
	markdownTableHeaders(t, "Resources", []string{"Name", "Type"})
}

func validateInputsTableHeaders(t *testing.T) {
	markdownTableHeaders(t, "Inputs", []string{"Name", "Description", "Type", "Required"})
}

func validateOutputsTableHeaders(t *testing.T) {
	markdownTableHeaders(t, "Outputs", []string{"Name", "Description"})
}

func markdownTableHeaders(t *testing.T, header string, columns []string) {
	readmePath := os.Getenv("README_PATH")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to load markdown file: %v", err)
	}

	contents := string(data)
	requiredHeaders := []string{"## " + header}

	for _, requiredHeader := range requiredHeaders {
		headerPattern := regexp.MustCompile("(?m)^" + regexp.QuoteMeta(requiredHeader) + "\\s*$")
		headerLoc := headerPattern.FindStringIndex(contents)
		if headerLoc == nil {
			t.Errorf("Failed: README.md does not contain required header: %s", requiredHeader)
		} else {
			t.Logf("Success: README.md contains required header: %s", requiredHeader)
		}

		// Look for a table immediately after the header
		tablePattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(requiredHeader) + `(\s*\|.*\|)+\s*`)
		tableLoc := tablePattern.FindStringIndex(contents)
		if tableLoc == nil {
			t.Errorf("Failed: README.md does not contain a table immediately after the header: %s", requiredHeader)
		} else {
			t.Logf("Success: README.md contains a table immediately after the header: %s", requiredHeader)
		}

		// Check the table headers
		columnHeaders := strings.Join(columns, " \\| ")
		headerRowPattern := regexp.MustCompile(`(?m)\| ` + columnHeaders + ` \|`)
		headerRowLoc := headerRowPattern.FindStringIndex(contents[tableLoc[0]:tableLoc[1]])
		if headerRowLoc == nil {
			t.Errorf("Failed: README.md does not contain the correct headers in the table after: %s", requiredHeader)
		} else {
			t.Logf("Success: README.md contains the correct headers in the table after: %s", requiredHeader)
		}
	}
}
