package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"mvdan.cc/xurls/v2"
)

type ErrorResponse struct {
	Errors []ErrorDetail `json:"errors"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestMarkdown(t *testing.T) {
    t.Run("URLs", validateURLs)
    t.Run("Headers", validateReadmeHeaders)
    t.Run("NotEmpty", validateReadmeNotEmpty)
    t.Run("ResourceTableHeaders", validateResourceTableHeaders)
    t.Run("InputsTableHeaders", validateInputsTableHeaders)
    t.Run("OutputsTableHeaders", validateOutputsTableHeaders)
}

func checkRegistryURL(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var errorResponse ErrorResponse
	err = json.Unmarshal(body, &errorResponse)
	if err != nil {
		return false, err
	}

	for _, errorDetail := range errorResponse.Errors {
		if errorDetail.Code == "NAME_UNKNOWN" {
			return false, nil
		}
	}
	return true, nil
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
        wg.Add(1)
        go func(link string) {
            defer wg.Done()

            if strings.Contains(link, "registry.terraform.io/providers/") {
                isValid, err := checkRegistryURL(link)
                if err != nil || !isValid {
                    t.Errorf("Failed: Invalid registry URL: %s", link)
                    return
                }
            } else {
                resp, err := http.Get(link)
                if err != nil {
                    t.Errorf("Failed: URL: %s, Error: %v", link, err)
                    return
                }
                defer resp.Body.Close()

                if resp.StatusCode != http.StatusOK {
                    t.Errorf("Failed: URL: %s, Status code: %d", link, resp.StatusCode)
                } else {
                    t.Logf("Success: URL: %s, Status code: %d", link, resp.StatusCode)
                }
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
		"## Usage":     1,
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
			t.Errorf("Failed: README.md does not contain required header")
		} else {
			t.Logf("Success: README.md contains required header")
		}

		// Look for a table immediately after the header
		tablePattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(requiredHeader) + `(\s*\|.*\|)+\s*`)
		tableLoc := tablePattern.FindStringIndex(contents)
		if tableLoc == nil {
			t.Errorf("Failed: README.md does not contain a table immediately after the header")
		} else {
			t.Logf("Success: README.md contains a table immediately after the header")
		}

		// Check the table headers
		columnHeaders := strings.Join(columns, " \\| ")
		headerRowPattern := regexp.MustCompile(`(?m)\| ` + columnHeaders + ` \|`)
		headerRowLoc := headerRowPattern.FindStringIndex(contents[tableLoc[0]:tableLoc[1]])
		if headerRowLoc == nil {
			t.Errorf("Failed: README.md does not contain the correct column names in the table")
		} else {
			t.Logf("Success: README.md contains the correct column names in the table")
		}
	}
}
