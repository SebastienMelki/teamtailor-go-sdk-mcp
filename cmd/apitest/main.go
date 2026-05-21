// Command apitest fetches every candidate that has applied to a specific
// Teamtailor job (position), prints their name + email, and saves each
// candidate's resume into a local folder.
//
//	# .env (gitignored)
//	TEAMTAILOR_API_KEY=...
//	TEAMTAILOR_REGION=eu|na|au           # default: eu
//	TEAMTAILOR_API_VERSION=20240904      # default: 20240904
//
//	make apitest ARGS="-job 12345"
//	# or
//	go run ./cmd/apitest -job 12345 -out ./resumes
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"

	teamtailorv1 "github.com/sebastienmelki/teamtailor-go-sdk-mcp/internal/gen/teamtailor/v1"
	"github.com/sebastienmelki/teamtailor-go-sdk-mcp/pkg/teamtailor"
)

func main() {
	_ = godotenv.Load()

	jobID := flag.String("job", "", "Job (position) id to fetch candidates for — required")
	outDir := flag.String("out", "resumes", "Directory to save resumes into")
	pageSize := flag.Int("page-size", 30, "Page size when listing job applications (Teamtailor caps at 30)")
	dump := flag.Bool("dump", false, "Dump outgoing HTTP requests (and response status) to stderr")
	flag.Parse()

	if *jobID == "" {
		log.Fatal("-job <id> is required (the Teamtailor job/position id)")
	}

	apiKey := os.Getenv("TEAMTAILOR_API_KEY")
	if apiKey == "" {
		log.Fatal("TEAMTAILOR_API_KEY must be set (export it or put it in a .env)")
	}

	region := teamtailor.Region(envOr("TEAMTAILOR_REGION", string(teamtailor.RegionEU)))
	apiVersion := envOr("TEAMTAILOR_API_VERSION", teamtailor.DefaultAPIVersion)

	httpClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: &dumpingTransport{enabled: *dump, base: http.DefaultTransport},
	}

	client := teamtailor.NewClient(apiKey,
		teamtailor.WithRegion(region),
		teamtailor.WithAPIVersion(apiVersion),
		teamtailor.WithHTTPClient(httpClient),
	)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("create output dir %q: %v", *outDir, err)
	}

	fmt.Printf("Teamtailor: job=%s region=%s api-version=%s out=%s\n",
		*jobID, region, apiVersion, *outDir)
	fmt.Println(strings.Repeat("=", 70))

	ctx := context.Background()
	var s stats

	for page := int32(1); ; page++ {
		pageCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		apps, err := client.ListJobApplications(pageCtx, &teamtailorv1.ListJobApplicationsRequest{
			FilterJob:  *jobID,
			PageSize:   int32(*pageSize),
			PageNumber: page,
			Include:    "candidate",
		})
		cancel()
		if err != nil {
			log.Fatalf("ListJobApplications (page %d): %v", page, err)
		}
		if len(apps.GetData()) == 0 {
			if page == 1 {
				fmt.Println("(no applications found for this job)")
			}
			break
		}

		for _, app := range apps.GetData() {
			s.seen++
			candidateID := app.GetRelationships().GetCandidate().GetData().GetId()
			if candidateID == "" {
				fmt.Printf("  [skip] application %s — no candidate id in relationship\n", app.GetId())
				continue
			}
			handleCandidate(ctx, client, httpClient, candidateID, *outDir, &s)
		}

		if apps.GetLinks().GetNext() == "" {
			break
		}
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Candidates seen: %d   resumes saved: %d   no resume: %d   download errors: %d\n",
		s.seen, s.downloaded, s.noResume, s.dlErr)
	fmt.Printf("Output folder: %s\n", *outDir)
}

type stats struct {
	seen       int
	downloaded int
	noResume   int
	dlErr      int
}

func handleCandidate(
	ctx context.Context,
	client *teamtailor.Client,
	httpClient *http.Client,
	candidateID, outDir string,
	s *stats,
) {
	candCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cand, err := client.GetCandidate(candCtx, &teamtailorv1.GetCandidateRequest{Id: candidateID})
	if err != nil {
		fmt.Printf("  [err]  candidate %s — GetCandidate failed: %v\n", candidateID, err)
		s.dlErr++
		return
	}
	a := cand.GetData().GetAttributes()
	name := strings.TrimSpace(a.GetFirstName() + " " + a.GetLastName())
	if name == "" {
		name = "(no name)"
	}

	resumeURL := a.GetResume()
	if resumeURL == "" {
		resumeURL = a.GetOriginalResume()
	}
	if resumeURL == "" {
		fmt.Printf("  %-40s <%s>  — no resume on file\n", truncate(name, 40), a.GetEmail())
		s.noResume++
		return
	}

	path := filepath.Join(outDir, resumeFilename(name, candidateID, resumeURL))
	if err := downloadFile(candCtx, httpClient, resumeURL, path); err != nil {
		fmt.Printf("  [err]  %-34s <%s>  — download failed: %v\n",
			truncate(name, 34), a.GetEmail(), err)
		s.dlErr++
		return
	}
	fmt.Printf("  %-40s <%s>  → %s\n", truncate(name, 40), a.GetEmail(), path)
	s.downloaded++
}

func resumeFilename(name, candidateID, resumeURL string) string {
	base := sanitize(name)
	if base == "" {
		base = "candidate"
	}
	return fmt.Sprintf("%s_%s%s", base, candidateID, extFromURL(resumeURL))
}

// sanitize keeps [A-Za-z0-9] verbatim, maps spaces/dashes/underscores to `_`,
// collapses repeats, and trims leading/trailing underscores.
func sanitize(s string) string {
	var b strings.Builder
	prevUnderscore := true
	for _, r := range s {
		switch {
		case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z', '0' <= r && r <= '9':
			b.WriteRune(r)
			prevUnderscore = false
		case r == ' ', r == '-', r == '_':
			if !prevUnderscore {
				b.WriteRune('_')
				prevUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

// extFromURL returns the file extension (with leading `.`) of the URL's last
// path segment, or ".pdf" if nothing recognizable is present.
func extFromURL(u string) string {
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	if i := strings.LastIndex(u, "/"); i >= 0 {
		u = u[i+1:]
	}
	if i := strings.LastIndex(u, "."); i >= 0 && len(u)-i <= 6 {
		return strings.ToLower(u[i:])
	}
	return ".pdf"
}

// downloadFile streams the body of `url` into `path`. It uses the supplied
// http.Client (so -dump still applies) but does NOT carry the Teamtailor
// Authorization header — resume URLs are public S3 / CDN links.
func downloadFile(ctx context.Context, client *http.Client, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// dumpingTransport prints each outgoing request (and the response status
// line) to stderr when enabled.
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
