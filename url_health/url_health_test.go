package urlhealth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Helper function to create a temporary file containing our test URLs
func createTempFile(t *testing.T, urls []string) string {
	file, err := os.CreateTemp("", "urls-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	for _, u := range urls {
		fmt.Fprintln(file, u)
	}
	file.Close()
	return file.Name()
}

func TestUrlHealth_InvalidWorkers(t *testing.T) {
	// Edge Case 1: Workers <= 0
	_, err := url_health("dummy.txt", 0)
	if err == nil {
		t.Error("Expected an error when workers <= 0, got nil")
	}
}

func TestUrlHealth_FileNotFound(t *testing.T) {
	// Edge Case 2: File doesn't exist
	_, err := url_health("does_not_exist_at_all.txt", 5)
	if err == nil {
		t.Error("Expected an error when file does not exist, got nil")
	}
}

func TestUrlHealth_EmptyFile(t *testing.T) {
	// Edge Case 3: File exists but has 0 URLs
	tempFile := createTempFile(t, []string{})
	defer os.Remove(tempFile)

	results, err := url_health(tempFile, 5)
	if err != nil {
		t.Errorf("Did not expect error for empty file, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestUrlHealth_CoreLogic(t *testing.T) {
	// Spin up a fake server to simulate different network conditions
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.WriteHeader(http.StatusOK)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/timeout":
			// Sleep for 3 seconds to trigger our worker's 2-second context timeout
			time.Sleep(3 * time.Second)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	// Prepare our test cases
	testURLs := []string{
		ts.URL + "/success",       // Edge Case 4: Perfect 200 OK
		ts.URL + "/notfound",      // Edge Case 5: Valid URL, but returns 404
		ts.URL + "/timeout",       // Edge Case 6: Server hangs, context timeout must catch it
		"htt://malformed-url-%%%", // Edge Case 7: Completely broken URL string
	}

	tempFile := createTempFile(t, testURLs)
	defer os.Remove(tempFile)

	// Run the function with 2 workers to force concurrency
	results, err := url_health(tempFile, 2)
	if err != nil {
		t.Fatalf("Unexpected error running url_health: %v", err)
	}

	// We expect exactly 4 results back, though order isn't guaranteed due to concurrency
	if len(results) != len(testURLs) {
		t.Fatalf("Expected %d results, got %d", len(testURLs), len(results))
	}

	// Validate the specifics of the edge cases
	successCount := 0
	timeoutCount := 0
	malformedCount := 0
	notFoundCount := 0

	for _, res := range results {
		if res.Url == ts.URL+"/success" {
			if !res.Success || res.Status_code != 200 {
				t.Errorf("/success failed: %+v", res)
			} else {
				successCount++
			}
		}

		if res.Url == ts.URL+"/notfound" {
			// A 404 is a successful HTTP request, it just returned a 404 status.
			if !res.Success || res.Status_code != 404 {
				t.Errorf("/notfound failed: %+v", res)
			} else {
				notFoundCount++
			}
		}

		if res.Url == ts.URL+"/timeout" {
			// Context timeout should result in an error and Success = false
			if res.Success || res.Error == nil {
				t.Errorf("/timeout should have failed due to context timeout: %+v", res)
			} else {
				timeoutCount++
			}
		}

		if res.Url == "htt://malformed-url-%%%" {
			// Request parsing should fail before it even sends
			if res.Success || res.Error == nil {
				t.Errorf("malformed URL should have failed: %+v", res)
			} else {
				malformedCount++
			}
		}
	}

	// Ensure all specific conditions were met
	if successCount != 1 || notFoundCount != 1 || timeoutCount != 1 || malformedCount != 1 {
		t.Errorf("One or more edge cases were not caught correctly. Success: %d, NotFound: %d, Timeout: %d, Malformed: %d", 
			successCount, notFoundCount, timeoutCount, malformedCount)
	}
}