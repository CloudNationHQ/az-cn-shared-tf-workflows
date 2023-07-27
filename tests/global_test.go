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

func TestURLs(t *testing.T) {
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

func TestReadmeHeaders(t *testing.T) {
	readmePath := os.Getenv("README_PATH")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to load markdown file: %v", err)
	}

	contents := string(data)
	requiredHeaders := []string{"## Goals", "## Features"}

	for _, header := range requiredHeaders {
		match, _ := regexp.MatchString("(?m)^"+regexp.QuoteMeta(header)+"$", contents)
		if !match {
			t.Errorf("Failed: README.md does not contain required header: %s", header)
		} else {
			t.Logf("Success: README.md contains required header: %s", header)
		}
	}
}

func TestReadmeNotEmpty(t *testing.T) {
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
