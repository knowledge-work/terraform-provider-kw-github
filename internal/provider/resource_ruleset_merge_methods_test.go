package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/knowledge-work/terraform-provider-kw-github/internal/githubclient"
)

// Mock GitHub API tests
func TestResourceRulesetAllowedMergeMethodsWithMock(t *testing.T) {
	// Mock server setup
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Track API calls
	var getCalled, putCalled bool

	// Mock GET /repos/owner/repo/rulesets/123 - return simple response for testing
	mux.HandleFunc("/repos/owner/repo/rulesets/123", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Return minimal valid response that matches go-github expectations
			fmt.Fprint(w, `{
				"id": 123,
				"name": "test-ruleset",
				"target": "branch",
				"enforcement": "active"
			}`)
		case "PUT":
			putCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
				"id": 123,
				"name": "test-ruleset",
				"target": "branch",
				"enforcement": "active"
			}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Create GitHub client pointing to mock server
	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

	// Test upsert function directly
	t.Run("Upsert", func(t *testing.T) {
		resource := &rulesetAllowedMergeMethodsResource{
			client: &githubclient.Client{Client: client, Owner: "owner"},
		}

		plan := &rulesetAllowedMergeMethodsResourceModel{
			Repository:          types.StringValue("repo"),
			RulesetID:           types.StringValue("123"),
			AllowedMergeMethods: convertToSet([]string{"merge"}),
		}

		err := resource.upsert(context.Background(), plan)
		if err != nil {
			t.Fatalf("upsert failed: %v", err)
		}

		if !getCalled {
			t.Error("Expected GET request to be called")
		}
		if !putCalled {
			t.Error("Expected PUT request to be called")
		}
	})

	// Test Delete operation (simplified - just test the logic)
	t.Run("Delete", func(t *testing.T) {
		// Reset call tracking
		getCalled, putCalled = false, false

		resource := &rulesetAllowedMergeMethodsResource{
			client: &githubclient.Client{Client: client, Owner: "owner"},
		}

		// Test the delete logic by calling the same upsert function
		// that Delete would call internally (with default methods)
		plan := &rulesetAllowedMergeMethodsResourceModel{
			Repository:          types.StringValue("repo"),
			RulesetID:           types.StringValue("123"),
			AllowedMergeMethods: convertToSet([]string{"merge", "squash", "rebase"}), // default methods
		}

		err := resource.upsert(context.Background(), plan)
		if err != nil {
			t.Fatalf("Delete (via upsert) failed: %v", err)
		}

		if !getCalled {
			t.Error("Expected GET request to be called during delete")
		}
		if !putCalled {
			t.Error("Expected PUT request to be called during delete")
		}
	})

	// Test error handling
	t.Run("UpsertError", func(t *testing.T) {
		// Create a new server that returns errors
		errorMux := http.NewServeMux()
		errorServer := httptest.NewServer(errorMux)
		defer errorServer.Close()

		errorMux.HandleFunc("/repos/owner/repo/rulesets/123", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message": "Not Found"}`)
		})

		errorClient := github.NewClient(nil)
		errorClient.BaseURL, _ = errorClient.BaseURL.Parse(errorServer.URL + "/")

		resource := &rulesetAllowedMergeMethodsResource{
			client: &githubclient.Client{Client: errorClient},
		}

		plan := &rulesetAllowedMergeMethodsResourceModel{
			Repository:          types.StringValue("owner/repo"),
			RulesetID:           types.StringValue("123"),
			AllowedMergeMethods: convertToSet([]string{"merge"}),
		}

		err := resource.upsert(context.Background(), plan)
		if err == nil {
			t.Error("Expected upsert to fail with 404 error")
		}
	})
}

// Unit tests for helper functions
func TestParseID(t *testing.T) {
	tests := []struct {
		input       string
		expected    int64
		expectError bool
	}{
		{"123", 123, false},
		{"456789", 456789, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"", 0, true},
		{"123abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := parseID(test.input)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("Expected %d, got %d", test.expected, result)
				}
			}
		})
	}
}

func TestConvertToSet(t *testing.T) {
	methods := []string{"merge", "squash", "rebase"}
	set := convertToSet(methods)

	if set.IsNull() {
		t.Error("Expected non-null set")
	}

	if set.IsUnknown() {
		t.Error("Expected known set")
	}

	elements := set.Elements()
	if len(elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(elements))
	}
}

func TestExtractMethodsFromSet(t *testing.T) {
	methods := []string{"merge", "squash", "rebase"}
	set := convertToSet(methods)

	extracted := extractMethodsFromSet(set)
	if len(extracted) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(extracted))
	}

	// Check if all methods are present (order independent)
	methodMap := make(map[string]bool)
	for _, method := range extracted {
		methodMap[method] = true
	}

	for _, expected := range methods {
		if !methodMap[expected] {
			t.Errorf("Expected method %q not found in extracted methods", expected)
		}
	}
}

func TestMethodsEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"same order", []string{"merge", "squash"}, []string{"merge", "squash"}, true},
		{"different order", []string{"merge", "squash"}, []string{"squash", "merge"}, true},
		{"different length", []string{"merge"}, []string{"merge", "squash"}, false},
		{"different content", []string{"merge", "rebase"}, []string{"merge", "squash"}, false},
		{"empty slices", []string{}, []string{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := methodsEqual(test.a, test.b)
			if result != test.expected {
				t.Errorf("Expected %v, got %v for %v vs %v", test.expected, result, test.a, test.b)
			}
		})
	}
}
