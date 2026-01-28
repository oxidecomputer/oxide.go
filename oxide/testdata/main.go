// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build ignore

// This script records real API responses for use in golden file tests.
// Run with: go run ./oxide/testdata/main.go [-api-version VERSION]
//
// Requires OXIDE_HOST, OXIDE_TOKEN, and OXIDE_PROJECT environment variables.
// Optionally pass -api-version to set the API-Version header on requests.
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var (
	apiVersion = flag.String("api-version", "", "API version to send in requests (optional)")

	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func main() {
	flag.Parse()

	host := os.Getenv("OXIDE_HOST")
	token := os.Getenv("OXIDE_TOKEN")
	project := os.Getenv("OXIDE_PROJECT")

	if host == "" || token == "" || project == "" {
		log.Fatalf("OXIDE_HOST, OXIDE_TOKEN, and OXIDE_PROJECT environment variables must be set")
	}

	if *apiVersion != "" {
		fmt.Printf("Using API-Version: %s\n", *apiVersion)
	}

	testdataDir := "./oxide/testdata/recordings"

	recordTimeseriesQuery(host, token, testdataDir)
	recordDiskList(host, token, project, testdataDir)
	recordLoopbackAddresses(host, token, testdataDir)
	recordIpPoolRanges(host, token, testdataDir)
}

func recordTimeseriesQuery(host, token, testdataDir string) {
	fmt.Println("Recording timeseries query response...")

	body := `{"query": "get hardware_component:voltage | filter slot == 0 && sensor == \"V1P0_MGMT\" | filter timestamp > @now() - 5m | last 5"}`
	data, err := doRequest("POST", host+"/v1/system/timeseries/query", token, body)
	if err != nil {
		log.Printf("Warning: timeseries query failed: %v", err)
		return
	}

	normalized, err := normalizeJSON(data)
	if err != nil {
		log.Printf("Warning: failed to normalize JSON: %v", err)
		return
	}
	if err := saveFixture(testdataDir, "timeseries_query_response.json", normalized); err != nil {
		log.Printf("Warning: %v", err)
		return
	}
}

func recordDiskList(host, token, project, testdataDir string) {
	fmt.Println("Recording disk list response...")

	url := fmt.Sprintf("%s/v1/disks?project=%s&limit=5", host, project)
	data, err := doRequest("GET", url, token, "")
	if err != nil {
		log.Printf("Warning: disk list failed: %v", err)
		return
	}

	normalized, err := normalizeJSON(data)
	if err != nil {
		log.Printf("Warning: failed to normalize JSON: %v", err)
		return
	}
	if err := saveFixture(testdataDir, "disk_list_response.json", normalized); err != nil {
		log.Printf("Warning: %v", err)
		return
	}
}

func recordLoopbackAddresses(host, token, testdataDir string) {
	fmt.Println("Recording loopback addresses response...")

	url := fmt.Sprintf("%s/v1/system/networking/loopback-address?limit=5", host)
	data, err := doRequest("GET", url, token, "")
	if err != nil {
		log.Printf("Warning: loopback addresses failed: %v", err)
		return
	}

	normalized, err := normalizeJSON(data)
	if err != nil {
		log.Printf("Warning: failed to normalize JSON: %v", err)
		return
	}
	if err := saveFixture(testdataDir, "loopback_addresses_response.json", normalized); err != nil {
		log.Printf("Warning: %v", err)
		return
	}
}

func recordIpPoolRanges(host, token, testdataDir string) {
	fmt.Println("Recording IP pool ranges response...")

	// Fetch ranges from specific pools to get both IPv4 and IPv6 coverage
	pools := []string{"fake-address", "fake-address-v6"}
	var allItems []any

	for _, poolName := range pools {
		url := fmt.Sprintf("%s/v1/system/ip-pools/%s/ranges?limit=1", host, poolName)
		data, err := doRequest("GET", url, token, "")
		if err != nil {
			log.Printf("Warning: IP pool ranges for %s failed: %v", poolName, err)
			continue
		}

		var resp struct {
			Items []any `json:"items"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Printf("Warning: failed to parse IP pool ranges response for %s: %v", poolName, err)
			continue
		}
		allItems = append(allItems, resp.Items...)
	}

	if len(allItems) == 0 {
		log.Printf("Warning: no IP pool ranges found, skipping recording")
		return
	}

	// Combine into a single response
	combined := map[string]any{
		"items": allItems,
	}
	data, err := json.Marshal(combined)
	if err != nil {
		log.Printf("Warning: failed to marshal combined response: %v", err)
		return
	}

	if err := saveFixture(testdataDir, "ip_pool_range_list_response.json", data); err != nil {
		log.Printf("Warning: %v", err)
		return
	}
}

// doRequest makes a request to the configured nexus instance. We use the standard library here
// and not our own sdk because we're generating test files to verify the generated code.
func doRequest(method, url, token, body string) ([]byte, error) {
	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if *apiVersion != "" {
		req.Header.Set("API-Version", *apiVersion)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, respBody)
	}

	return io.ReadAll(resp.Body)
}

func saveFixture(testdataDir, filename string, data []byte) error {
	path := filepath.Join(testdataDir, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

// normalizeJSON strips undocumented fields from API responses.
func normalizeJSON(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}

	// Nexus returns an undocumented `query_summaries` field that's not in the OpenAPI spec. Ignore it for now.
	//
	// TODO: fully drop `query_summaries` from nexus unless requested.
	if m, ok := v.(map[string]any); ok {
		delete(m, "query_summaries")
	}

	return json.Marshal(v)
}
