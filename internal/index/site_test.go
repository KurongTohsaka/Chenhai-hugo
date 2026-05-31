package index_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/config"
	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
)

// Helper: creates a minimal Page for testing.
func testPage(title string, date time.Time, draft bool, categories, tags []string) *content.Page {
	return &content.Page{
		Title:      title,
		Date:       date,
		Draft:      draft,
		Categories: categories,
		Tags:       tags,
		RawContent: "Content for " + title,
		Summary:    "Summary of " + title,
		RelPath:    "posts/" + title + ".md",
		Section:    "posts",
	}
}

func TestBuildSite_Categories(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("post-a", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), false, []string{"Tech", "Go"}, nil),
		testPage("post-b", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, []string{"Tech", "Rust"}, nil),
		testPage("post-c", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, []string{"Life"}, nil),
		testPage("post-d", time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
	}

	site := index.BuildSite(cfg, pages, false)

	if len(site.Categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(site.Categories))
	}

	// Verify category keys
	if _, ok := site.Categories["Tech/Go"]; !ok {
		t.Errorf("expected category 'Tech/Go' to exist")
	}
	if _, ok := site.Categories["Tech/Rust"]; !ok {
		t.Errorf("expected category 'Tech/Rust' to exist")
	}
	if _, ok := site.Categories["Life"]; !ok {
		t.Errorf("expected category 'Life' to exist")
	}

	// Verify pages are in correct categories
	if len(site.Categories["Tech/Go"]) != 1 {
		t.Errorf("expected 1 page in 'Tech/Go', got %d", len(site.Categories["Tech/Go"]))
	}
	if site.Categories["Tech/Go"][0].Title != "post-a" {
		t.Errorf("expected 'post-a' in 'Tech/Go', got %s", site.Categories["Tech/Go"][0].Title)
	}

	// Page with no categories should not appear
	if _, ok := site.Categories[""]; ok {
		t.Error("page with no categories should not create a category entry")
	}
}

func TestBuildSite_Tags(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("post-a", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"go", "web"}),
		testPage("post-b", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"go", "database"}),
		testPage("post-c", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"web", "frontend"}),
	}

	site := index.BuildSite(cfg, pages, false)

	if len(site.Tags) != 4 {
		t.Fatalf("expected 4 tags, got %d", len(site.Tags))
	}

	// Verify tag counts - "go" appears on 2 pages
	if len(site.Tags["go"]) != 2 {
		t.Errorf("expected 2 pages for tag 'go', got %d", len(site.Tags["go"]))
	}
	// "web" appears on 2 pages
	if len(site.Tags["web"]) != 2 {
		t.Errorf("expected 2 pages for tag 'web', got %d", len(site.Tags["web"]))
	}
	// "database" appears on 1 page
	if len(site.Tags["database"]) != 1 {
		t.Errorf("expected 1 page for tag 'database', got %d", len(site.Tags["database"]))
	}
	// "frontend" appears on 1 page
	if len(site.Tags["frontend"]) != 1 {
		t.Errorf("expected 1 page for tag 'frontend', got %d", len(site.Tags["frontend"]))
	}
}

func TestBuildSite_Archives(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("jan2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("feb2024", time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("mar2024", time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("dec2023", time.Date(2023, 12, 20, 0, 0, 0, 0, time.UTC), false, nil, nil),
	}

	site := index.BuildSite(cfg, pages, false)

	// Should have 2 years
	if len(site.Archives.Years) != 2 {
		t.Fatalf("expected 2 years in archives, got %v", site.Archives.Years)
	}

	// Check year 2024 has 3 months
	if len(site.Archives.Items[2024]) != 3 {
		t.Errorf("expected 3 months in 2024, got %d", len(site.Archives.Items[2024]))
	}

	// Check year 2023 has 1 month
	if len(site.Archives.Items[2023]) != 1 {
		t.Errorf("expected 1 month in 2023, got %d", len(site.Archives.Items[2023]))
	}

	// Verify specific month contents
	for _, m := range site.Archives.Items[2024] {
		if m.Month == time.January {
			if len(m.Pages) != 1 {
				t.Fatalf("expected 1 page in Jan 2024, got %d", len(m.Pages))
			}
			if m.Pages[0].Title != "jan2024" {
				t.Errorf("expected 'jan2024' in Jan 2024, got %s", m.Pages[0].Title)
			}
		}
	}

	for _, m := range site.Archives.Items[2023] {
		if m.Month == time.December {
			if len(m.Pages) != 1 {
				t.Fatalf("expected 1 page in Dec 2023, got %d", len(m.Pages))
			}
			if m.Pages[0].Title != "dec2023" {
				t.Errorf("expected 'dec2023' in Dec 2023, got %s", m.Pages[0].Title)
			}
		}
	}
}

func TestBuildSite_DraftExcluded(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("draft-post", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), true, []string{"Tech"}, []string{"draft-tag"}),
		testPage("published-post", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, []string{"Tech"}, []string{"real-tag"}),
	}

	site := index.BuildSite(cfg, pages, false)

	// Draft page should not appear in Categories
	if _, ok := site.Categories["Tech"]; !ok {
		t.Fatal("expected 'Tech' category to exist")
	}
	if len(site.Categories["Tech"]) != 1 {
		t.Errorf("expected 1 page in 'Tech' category (draft excluded), got %d", len(site.Categories["Tech"]))
	}
	if site.Categories["Tech"][0].Title != "published-post" {
		t.Errorf("expected 'published-post' in category, got %s", site.Categories["Tech"][0].Title)
	}

	// Draft page should not appear in Tags
	if _, ok := site.Tags["real-tag"]; !ok {
		t.Fatal("expected 'real-tag' tag to exist")
	}
	if _, ok := site.Tags["draft-tag"]; ok {
		t.Error("draft page tag 'draft-tag' should not appear in Tags index")
	}

	// Draft page should not appear in Archives
	totalArchivePages := 0
	for _, months := range site.Archives.Items {
		for _, m := range months {
			totalArchivePages += len(m.Pages)
		}
	}
	if totalArchivePages != 1 {
		t.Errorf("expected 1 page in archives (draft excluded), got %d", totalArchivePages)
	}

	// Draft page should NOT be in the Pages slice when showDrafts=false
	if len(site.Pages) != 1 {
		t.Errorf("expected 1 page in total (draft excluded from Pages when showDrafts=false), got %d", len(site.Pages))
	}
}

func TestBuildSite_SortOrder(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("oldest", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("middle", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("newest", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
	}

	site := index.BuildSite(cfg, pages, false)

	if len(site.Pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(site.Pages))
	}

	// Verify descending sort
	if site.Pages[0].Title != "newest" {
		t.Errorf("expected first page to be 'newest', got %s", site.Pages[0].Title)
	}
	if site.Pages[1].Title != "middle" {
		t.Errorf("expected second page to be 'middle', got %s", site.Pages[1].Title)
	}
	if site.Pages[2].Title != "oldest" {
		t.Errorf("expected third page to be 'oldest', got %s", site.Pages[2].Title)
	}
}

func TestBuildSearchIndex(t *testing.T) {
	pages := []*content.Page{
		testPage("searchable", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), false, []string{"Tech"}, []string{"go", "search"}),
		testPage("draft-hidden", time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), true, nil, nil),
	}

	data, err := index.BuildSearchIndex(pages)
	if err != nil {
		t.Fatalf("BuildSearchIndex failed: %v", err)
	}

	// Verify it's valid JSON
	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// Only non-draft pages should be in the index
	if len(entries) != 1 {
		t.Fatalf("expected 1 search entry (draft excluded), got %d", len(entries))
	}

	entry := entries[0]
	if entry["title"] != "searchable" {
		t.Errorf("expected title 'searchable', got %v", entry["title"])
	}
	if entry["tags"] == nil {
		t.Error("expected tags to be present")
	}
	if entry["category"] != "Tech" {
		t.Errorf("expected category 'Tech', got %v", entry["category"])
	}
	if entry["date"] != "2024-06-15" {
		t.Errorf("expected date '2024-06-15', got %v", entry["date"])
	}
	if entry["summary"] != "Summary of searchable" {
		t.Errorf("expected summary 'Summary of searchable', got %v", entry["summary"])
	}
	if entry["content"] == nil {
		t.Error("expected content to be present")
	}
}

func TestBuildSearchIndex_ContentTruncation(t *testing.T) {
	longContent := strings.Repeat("a", 1000)
	page := &content.Page{
		Title:      "long-post",
		Date:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Draft:      false,
		RawContent: longContent,
		RelPath:    "posts/long-post.md",
		Section:    "posts",
	}

	data, err := index.BuildSearchIndex([]*content.Page{page})
	if err != nil {
		t.Fatalf("BuildSearchIndex failed: %v", err)
	}

	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	contentStr := entries[0]["content"].(string)
	if len([]rune(contentStr)) > 503 {
		t.Errorf("content should be truncated to ~500 runes + '...', got %d runes", len([]rune(contentStr)))
	}
	if !strings.HasSuffix(contentStr, "...") {
		t.Error("truncated content should end with '...'")
	}
}

func TestBuildSearchIndex_DraftExcluded(t *testing.T) {
	pages := []*content.Page{
		testPage("draft", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true, nil, nil),
	}
	data, err := index.BuildSearchIndex(pages)
	if err != nil {
		t.Fatalf("BuildSearchIndex failed: %v", err)
	}

	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries for only draft pages, got %d", len(entries))
	}
}

func TestBuildTagCloud(t *testing.T) {
	cfg := config.DefaultConfig()
	// Create pages with different tag frequencies
	pages := []*content.Page{
		testPage("p1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"rare"}),           // 1 page
		testPage("p2", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"rare"}),           // still 1 for rare
		testPage("p3", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"popular"}),        // 1 for popular
		testPage("p4", time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"popular"}),        // 2
		testPage("p5", time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"popular"}),        // 3
		testPage("p6", time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"common", "rare"}), // common=1, rare=2
		testPage("p7", time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"common"}),         // common=2
		testPage("p8", time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"common"}),         // common=3
		testPage("p9", time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"common"}),         // common=4
		testPage("p10", time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"popular"}),      // popular=4
		testPage("p11", time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"popular"}),      // popular=5
	}

	site := index.BuildSite(cfg, pages, false)
	cloud := site.BuildTagCloud()

	// Verify all 3 tags present
	if len(cloud) != 3 {
		t.Fatalf("expected 3 tag cloud entries, got %d", len(cloud))
	}

	// Tags should be sorted alphabetically: "common", "popular", "rare"
	if cloud[0].Name != "common" || cloud[1].Name != "popular" || cloud[2].Name != "rare" {
		t.Errorf("tags not sorted alphabetically: got %v", []string{cloud[0].Name, cloud[1].Name, cloud[2].Name})
	}

	// Verify counts
	if cloud[0].Count != 4 { // common -> 4
		t.Errorf("expected 'common' count 4, got %d", cloud[0].Count)
	}
	if cloud[1].Count != 5 { // popular -> 5
		t.Errorf("expected 'popular' count 5, got %d", cloud[1].Count)
	}
	// "rare" appears on p1, p2, and p6 = 3 pages
	if cloud[2].Count != 3 {
		t.Errorf("expected 'rare' count 3, got %d", cloud[2].Count)
	}

	// minCount=3, maxCount=5
	// common(4): idx = (4-3)*4/(5-3) = 4/2 = 2 -> "md"
	// popular(5): idx = (5-3)*4/(5-3) = 8/2 = 4 -> "xl"
	// rare(3):    idx = (3-3)*4/(5-3) = 0 -> "xs"
	sizes := map[string]string{
		"common":  "md",
		"popular": "xl",
		"rare":    "xs",
	}
	for _, entry := range cloud {
		if entry.Size != sizes[entry.Name] {
			t.Errorf("tag %q: expected size %q, got %q", entry.Name, sizes[entry.Name], entry.Size)
		}
	}
}

func TestBuildTagCloud_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	site := index.BuildSite(cfg, nil, false)
	cloud := site.BuildTagCloud()

	if cloud != nil {
		t.Errorf("expected nil tag cloud for empty site, got %v", cloud)
	}
}

func TestBuildTagCloud_NoTags(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("no-tags", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
	}
	site := index.BuildSite(cfg, pages, false)
	cloud := site.BuildTagCloud()

	if cloud != nil {
		t.Errorf("expected nil tag cloud for pages with no tags, got %v", cloud)
	}
}

func TestBuildTagCloud_SingleTag(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("p1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"only"}),
		testPage("p2", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"only"}),
	}
	site := index.BuildSite(cfg, pages, false)
	cloud := site.BuildTagCloud()

	if len(cloud) != 1 {
		t.Fatalf("expected 1 tag cloud entry, got %d", len(cloud))
	}

	// Single tag, minCount == maxCount, so idx should be 0 -> "xs"
	if cloud[0].Name != "only" {
		t.Errorf("expected tag name 'only', got %s", cloud[0].Name)
	}
	if cloud[0].Count != 2 {
		t.Errorf("expected count 2, got %d", cloud[0].Count)
	}
	if cloud[0].Size != "xs" {
		t.Errorf("expected size 'xs' for single tag (min==max), got %s", cloud[0].Size)
	}
}

func TestPublishedPages(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("draft-one", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), true, nil, nil),
		testPage("published", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
		testPage("draft-two", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true, nil, nil),
		testPage("also-published", time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), false, nil, nil),
	}

	site := index.BuildSite(cfg, pages, false)
	published := site.PublishedPages()

	if len(published) != 2 {
		t.Fatalf("expected 2 published pages, got %d", len(published))
	}

	// Published pages should still be in original date-descending order
	for _, p := range published {
		if p.Draft {
			t.Errorf("published page %q has Draft=true", p.Title)
		}
	}
}

func TestPagesByCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("a", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, []string{"X"}, nil),
		testPage("b", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, []string{"X"}, nil),
		testPage("c", time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), false, []string{"Y"}, nil),
	}

	site := index.BuildSite(cfg, pages, false)

	catX := site.PagesByCategory("X")
	if len(catX) != 2 {
		t.Errorf("expected 2 pages in category 'X', got %d", len(catX))
	}

	catY := site.PagesByCategory("Y")
	if len(catY) != 1 {
		t.Errorf("expected 1 page in category 'Y', got %d", len(catY))
	}

	catNone := site.PagesByCategory("Z")
	if len(catNone) != 0 {
		t.Errorf("expected 0 pages in category 'Z', got %d", len(catNone))
	}
}

func TestPagesByTag(t *testing.T) {
	cfg := config.DefaultConfig()
	pages := []*content.Page{
		testPage("a", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"tag1", "tag2"}),
		testPage("b", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), false, nil, []string{"tag1"}),
	}

	site := index.BuildSite(cfg, pages, false)

	tag1 := site.PagesByTag("tag1")
	if len(tag1) != 2 {
		t.Errorf("expected 2 pages for tag 'tag1', got %d", len(tag1))
	}

	tag2 := site.PagesByTag("tag2")
	if len(tag2) != 1 {
		t.Errorf("expected 1 page for tag 'tag2', got %d", len(tag2))
	}

	tagNone := site.PagesByTag("nonexistent")
	if len(tagNone) != 0 {
		t.Errorf("expected 0 pages for nonexistent tag, got %d", len(tagNone))
	}
}
