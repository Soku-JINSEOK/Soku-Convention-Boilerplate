package initcmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	maxArchiveBytes   = 64 << 20
	maxExtractedBytes = 128 << 20
	maxFileBytes      = 8 << 20
	maxArchiveEntries = 10000
)

type SourceClient struct {
	HTTP    *http.Client
	APIBase string
}

func NewSourceClient() *SourceClient {
	client := &http.Client{Timeout: 60 * time.Second}
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many redirects")
		}
		if request.URL.Scheme != "https" {
			return errors.New("source redirect is not HTTPS")
		}
		return nil
	}
	return &SourceClient{HTTP: client, APIBase: "https://api.github.com"}
}

func (client *SourceClient) Fetch(ctx context.Context, source, release string) (SourceSnapshot, error) {
	owner, repository, err := parseGitHubSource(source)
	if err != nil {
		return SourceSnapshot{}, err
	}
	if client.HTTP == nil {
		client.HTTP = NewSourceClient().HTTP
	}
	if client.APIBase == "" {
		client.APIBase = "https://api.github.com"
	}
	var ref struct {
		Object struct{ Type, SHA, URL string } `json:"object"`
	}
	endpoint := fmt.Sprintf("%s/repos/%s/%s/git/ref/tags/%s", strings.TrimSuffix(client.APIBase, "/"), owner, repository, release)
	if err := client.getJSON(ctx, endpoint, &ref); err != nil {
		return SourceSnapshot{}, err
	}
	objectType, sha, objectURL := ref.Object.Type, strings.ToLower(ref.Object.SHA), ref.Object.URL
	for depth := 0; objectType == "tag"; depth++ {
		if depth >= 8 {
			return SourceSnapshot{}, fail(6, "source.invalid", "annotated tag chain is too deep")
		}
		var tag struct {
			Object struct{ Type, SHA, URL string } `json:"object"`
		}
		if objectURL == "" {
			objectURL = fmt.Sprintf("%s/repos/%s/%s/git/tags/%s", strings.TrimSuffix(client.APIBase, "/"), owner, repository, sha)
		}
		if !strings.HasPrefix(objectURL, strings.TrimSuffix(client.APIBase, "/")+"/") {
			return SourceSnapshot{}, fail(6, "source.invalid", "annotated tag object URL leaves the GitHub API origin")
		}
		if err := client.getJSON(ctx, objectURL, &tag); err != nil {
			return SourceSnapshot{}, err
		}
		objectType, sha, objectURL = tag.Object.Type, strings.ToLower(tag.Object.SHA), tag.Object.URL
	}
	if objectType != "commit" || !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(sha) {
		return SourceSnapshot{}, fail(6, "source.invalid", "release tag does not resolve to a full commit")
	}
	archiveURL := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", strings.TrimSuffix(client.APIBase, "/"), owner, repository, sha)
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, archiveURL, nil)
	client.authorize(request)
	response, err := client.HTTP.Do(request)
	if err != nil {
		return SourceSnapshot{}, fail(6, "source.fetch", "download source archive: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return SourceSnapshot{}, fail(6, "source.fetch", "download source archive: GitHub returned %s", response.Status)
	}
	if response.Request.URL.Scheme != "https" && !isLoopbackTestURL(response.Request.URL, client.APIBase) {
		return SourceSnapshot{}, fail(6, "source.fetch", "archive redirect target must use HTTPS")
	}
	limited := io.LimitReader(response.Body, maxArchiveBytes+1)
	archive, err := io.ReadAll(limited)
	if err != nil {
		return SourceSnapshot{}, fail(6, "source.fetch", "read source archive: %v", err)
	}
	if len(archive) > maxArchiveBytes {
		return SourceSnapshot{}, fail(6, "source.archive", "source archive exceeds the compressed size limit")
	}
	files, err := extractArchive(archive)
	if err != nil {
		return SourceSnapshot{}, err
	}
	return SourceSnapshot{Source: source, Release: release, ResolvedCommit: sha, Files: files}, nil
}

func parseGitHubSource(value string) (string, string, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || !strings.EqualFold(parsed.Host, "github.com") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", "", fail(2, "source.invalid", "boilerplate_source must be https://github.com/<owner>/<repo> without credentials, query, or fragment")
	}
	parts := strings.Split(strings.Trim(parsed.EscapedPath(), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[1], ".git") {
		return "", "", fail(2, "source.invalid", "boilerplate_source must be https://github.com/<owner>/<repo>")
	}
	for _, part := range parts {
		if !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(part) {
			return "", "", fail(2, "source.invalid", "boilerplate_source contains an invalid repository component")
		}
	}
	return parts[0], parts[1], nil
}

func (client *SourceClient) getJSON(ctx context.Context, endpoint string, target any) error {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	client.authorize(request)
	response, err := client.HTTP.Do(request)
	if err != nil {
		return fail(6, "source.fetch", "resolve release: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fail(6, "source.fetch", "resolve release: GitHub returned %s", response.Status)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return fail(6, "source.fetch", "decode GitHub response: %v", err)
	}
	return nil
}

func (client *SourceClient) authorize(request *http.Request) {
	token := os.Getenv("GH_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
}

func extractArchive(data []byte) (map[string][]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fail(6, "source.archive", "source archive is not a valid gzip stream")
	}
	defer func() { _ = gzipReader.Close() }()
	reader := tar.NewReader(gzipReader)
	files := map[string][]byte{}
	folded := map[string]string{}
	var prefix string
	var total int64
	for count := 0; ; count++ {
		if count >= maxArchiveEntries {
			return nil, fail(6, "source.archive", "source archive exceeds the entry limit")
		}
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fail(6, "source.archive", "read source archive: %v", err)
		}
		// GitHub source archives begin with a POSIX PAX global metadata entry.
		// It describes later headers and is not part of the archive's file root.
		if header.Typeflag == tar.TypeXGlobalHeader {
			continue
		}
		rawName := strings.ReplaceAll(header.Name, "\\", "/")
		if strings.HasPrefix(rawName, "/") {
			return nil, fail(6, "source.archive", "source archive contains an unsafe path")
		}
		for _, component := range strings.Split(rawName, "/") {
			if component == ".." {
				return nil, fail(6, "source.archive", "source archive contains traversal")
			}
		}
		clean := path.Clean(rawName)
		if strings.HasPrefix(clean, "/") || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
			return nil, fail(6, "source.archive", "source archive contains an unsafe path")
		}
		parts := strings.Split(clean, "/")
		if prefix == "" {
			prefix = parts[0]
		}
		if parts[0] != prefix {
			return nil, fail(6, "source.archive", "source archive has multiple roots")
		}
		if len(parts) == 1 {
			continue
		}
		relative := strings.Join(parts[1:], "/")
		if err := validateArchivePath(relative); err != nil {
			return nil, err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
		default:
			return nil, fail(6, "source.archive", "source archive contains a link, device, or unsupported entry")
		}
		if header.Size < 0 || header.Size > maxFileBytes {
			return nil, fail(6, "source.archive", "source archive file %q exceeds the size limit", relative)
		}
		total += header.Size
		if total > maxExtractedBytes {
			return nil, fail(6, "source.archive", "source archive exceeds the extracted size limit")
		}
		lower := strings.ToLower(relative)
		if previous, exists := folded[lower]; exists {
			return nil, fail(6, "source.archive", "source archive paths %q and %q collide by case", previous, relative)
		}
		folded[lower] = relative
		content, err := io.ReadAll(io.LimitReader(reader, maxFileBytes+1))
		if err != nil || int64(len(content)) != header.Size {
			return nil, fail(6, "source.archive", "read source archive file %q", relative)
		}
		if shouldScanArchiveSecret(relative) && secretBearing(content) {
			return nil, fail(6, "source.archive", "source archive file %q appears to contain a secret", relative)
		}
		files[relative] = content
	}
	if _, ok := files[CatalogPath]; !ok {
		return nil, fail(5, "catalog.incompatible", "source release does not contain %s", CatalogPath)
	}
	return files, nil
}

func validateArchivePath(value string) error {
	if value == "" || strings.Contains(value, "\\") || path.Clean(value) != value {
		return fail(6, "source.archive", "source archive contains a non-canonical path")
	}
	for _, component := range strings.Split(value, "/") {
		lower := strings.ToLower(component)
		base := strings.SplitN(lower, ".", 2)[0]
		if component == ".." || strings.ContainsAny(component, `<>:"|?*`) || strings.HasSuffix(component, ".") || strings.HasSuffix(component, " ") || contains([]string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}, base) {
			return fail(6, "source.archive", "source archive path %q is not portable", value)
		}
	}
	return nil
}

var secretPatterns = []*regexp.Regexp{regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`), regexp.MustCompile(`(?im)^[ \t]*['"]?(?:api[_-]?key|access[_-]?token|client[_-]?secret|password)['"]?[ \t]*[:=][ \t]*['"]?[A-Za-z0-9_./+~-]{16,}`), regexp.MustCompile(`gh[pousr]_[A-Za-z0-9]{20,}`)}

func shouldScanArchiveSecret(name string) bool {
	// Manifest test fixtures deliberately contain forbidden configuration to
	// prove that the parser rejects it. They are never rendered downstream.
	return !strings.HasPrefix(name, "soku/testdata/")
}

func secretBearing(content []byte) bool {
	if bytes.IndexByte(content, 0) >= 0 {
		return false
	}
	for _, pattern := range secretPatterns {
		if pattern.Match(content) {
			return true
		}
	}
	return false
}
func isLoopbackTestURL(value *url.URL, apiBase string) bool {
	base, _ := url.Parse(apiBase)
	return base != nil && (base.Hostname() == "127.0.0.1" || base.Hostname() == "localhost") && value.Host == base.Host
}
