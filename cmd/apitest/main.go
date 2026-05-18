// Command apitest runs a manual integration harness against the live
// Teamtailor API.
//
//	# .env (optional)
//	TEAMTAILOR_API_KEY=...
//	TEAMTAILOR_REGION=eu|na|au           # default: eu
//	TEAMTAILOR_API_VERSION=20240904      # default: 20240904
//
//	make apitest
//
// Each test prints PASS/FAIL with a short note. With -dump the harness
// prints the outgoing HTTP request for each call (useful for verifying
// header + query-param encoding).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"

	teamtailorv1 "github.com/sebastienmelki/teamtailor-go-sdk-mcp/internal/gen/teamtailor/v1"
	"github.com/sebastienmelki/teamtailor-go-sdk-mcp/pkg/teamtailor"
)

type result struct {
	name    string
	ok      bool
	err     error
	details string
}

func main() {
	_ = godotenv.Load()

	dump := flag.Bool("dump", false, "Dump outgoing HTTP requests (and response status) to stderr")
	flag.Parse()

	apiKey := os.Getenv("TEAMTAILOR_API_KEY")
	if apiKey == "" {
		log.Fatal("TEAMTAILOR_API_KEY must be set (export it or put it in a .env)")
	}

	region := teamtailor.Region(envOr("TEAMTAILOR_REGION", string(teamtailor.RegionEU)))
	apiVersion := envOr("TEAMTAILOR_API_VERSION", teamtailor.DefaultAPIVersion)

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &dumpingTransport{enabled: *dump, base: http.DefaultTransport},
	}

	client := teamtailor.NewClient(apiKey,
		teamtailor.WithRegion(region),
		teamtailor.WithAPIVersion(apiVersion),
		teamtailor.WithHTTPClient(httpClient),
	)

	fmt.Printf("Teamtailor apitest — region=%s api-version=%s\n", region, apiVersion)
	fmt.Println(strings.Repeat("=", 70))

	results := []result{
		testListCandidates(client),
		testListCandidatesPagination(client),
	}

	printSummary(results)
}

func testListCandidates(c *teamtailor.Client) result {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := c.ListCandidates(ctx, &teamtailorv1.ListCandidatesRequest{PageSize: 5})
	if err != nil {
		return fail("ListCandidates (page-size=5)", err)
	}
	got := len(resp.GetData())
	details := fmt.Sprintf("got %d candidates", got)
	if got > 0 {
		first := resp.GetData()[0]
		details += fmt.Sprintf("; first=%s %s <%s>",
			first.GetAttributes().GetFirstName(),
			first.GetAttributes().GetLastName(),
			first.GetAttributes().GetEmail(),
		)
	}
	return pass("ListCandidates (page-size=5)", details)
}

func testListCandidatesPagination(c *teamtailor.Client) result {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	first, err := c.ListCandidates(ctx, &teamtailorv1.ListCandidatesRequest{PageSize: 2})
	if err != nil {
		return fail("ListCandidates pagination — first page", err)
	}
	nextURL := first.GetLinks().GetNext()
	if nextURL == "" {
		return result{
			name:    "ListCandidates pagination",
			ok:      true,
			details: "no next link (account has <= 2 candidates) — pagination not exercised",
		}
	}
	after := extractPageAfter(nextURL)
	if after == "" {
		return fail("ListCandidates pagination — parse page[after]",
			fmt.Errorf("next link missing page[after]: %s", nextURL))
	}
	second, err := c.ListCandidates(ctx, &teamtailorv1.ListCandidatesRequest{
		PageSize:  2,
		PageAfter: after,
	})
	if err != nil {
		return fail("ListCandidates pagination — second page", err)
	}
	return pass("ListCandidates pagination", fmt.Sprintf("page 2 returned %d candidates", len(second.GetData())))
}

// extractPageAfter pulls page[after]=... from a URL's query string.
func extractPageAfter(rawURL string) string {
	idx := strings.Index(rawURL, "page%5Bafter%5D=")
	if idx < 0 {
		idx = strings.Index(rawURL, "page[after]=")
		if idx < 0 {
			return ""
		}
		rest := rawURL[idx+len("page[after]="):]
		if amp := strings.IndexByte(rest, '&'); amp >= 0 {
			return rest[:amp]
		}
		return rest
	}
	rest := rawURL[idx+len("page%5Bafter%5D="):]
	if amp := strings.IndexByte(rest, '&'); amp >= 0 {
		return rest[:amp]
	}
	return rest
}

func pass(name, details string) result { return result{name: name, ok: true, details: details} }
func fail(name string, err error) result {
	return result{name: name, ok: false, err: err}
}

func printSummary(results []result) {
	fmt.Println(strings.Repeat("=", 70))
	passed, failed := 0, 0
	for _, r := range results {
		icon, status := "+", "PASS"
		if !r.ok {
			icon, status = "x", "FAIL"
			failed++
		} else {
			passed++
		}
		fmt.Printf("  [%s] %-45s %s", icon, r.name+":", status)
		if r.ok && r.details != "" {
			fmt.Printf(" — %s", r.details)
		}
		if !r.ok && r.err != nil {
			fmt.Printf(" — %s", r.err.Error())
		}
		fmt.Println()
	}
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Total: %d  Passed: %d  Failed: %d\n", len(results), passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// dumpingTransport prints each outgoing request (and the response status
// line) to stderr when enabled. It's a debugging aid for verifying that
// the generated client encodes headers and bracketed query params the
// way Teamtailor expects.
type dumpingTransport struct {
	enabled bool
	base    http.RoundTripper
}

func (d *dumpingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.enabled {
		dump, err := httputil.DumpRequestOut(req, false)
		if err == nil {
			fmt.Fprintf(os.Stderr, "\n--- HTTP request ---\n%s\n", string(dump))
		}
	}
	resp, err := d.base.RoundTrip(req)
	if d.enabled && resp != nil {
		fmt.Fprintf(os.Stderr, "--- HTTP response: %s ---\n\n", resp.Status)
	}
	return resp, err
}
