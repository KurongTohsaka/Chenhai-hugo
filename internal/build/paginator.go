package build

import (
	"fmt"
	"path/filepath"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
	"github.com/KurongTohsaka/chenhai-hugo/internal/index"
	"github.com/KurongTohsaka/chenhai-hugo/internal/theme"
)

// PageNum holds a single page number entry for the paginator.
type PageNum struct {
	Number     int
	IsCurrent  bool
	IsEllipsis bool
}

// Paginator holds pagination info for template rendering.
type Paginator struct {
	PageNumber  int
	TotalPages  int
	Pages       []*content.Page
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	PageNumbers []PageNum
}

// newPaginator creates a Paginator for a given page of paginated results.
func newPaginator(pages []*content.Page, page, perPage int) *Paginator {
	totalPages := (len(pages) + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	start := (page - 1) * perPage
	end := start + perPage
	if end > len(pages) {
		end = len(pages)
	}

	// Build page number list with ellipsis
	var nums []PageNum
	for i := 1; i <= totalPages; i++ {
		if i <= 3 || i > totalPages-3 || (i >= page-1 && i <= page+1) {
			nums = append(nums, PageNum{Number: i, IsCurrent: i == page})
		} else if len(nums) > 0 && !nums[len(nums)-1].IsEllipsis {
			nums = append(nums, PageNum{IsEllipsis: true})
		}
	}

	return &Paginator{
		PageNumber:  page,
		TotalPages:  totalPages,
		Pages:       pages[start:end],
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
		PageNumbers: nums,
	}
}

// renderPaginatedListPages renders paginated list pages for a given set of pages.
// baseDir is the directory for page 1; page/2/, page/3/ directories go under it.
func (b *Builder) renderPaginatedListPages(title string, pages []*content.Page, baseDir, tmpl string, site *index.Site, public string) error {
	perPage := b.cfg.ThemeConfig.PostsPerPage
	if perPage < 1 {
		perPage = 10
	}

	totalPages := (len(pages) + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	// Compute base URL path for pagination links (empty for homepage root)
	basePath := ""
	if baseDir != public {
		rel, _ := filepath.Rel(public, baseDir)
		basePath = "/" + filepath.ToSlash(rel)
	}

	for page := 1; page <= totalPages; page++ {
		paginator := newPaginator(pages, page, perPage)

		var outDir string
		if page == 1 {
			outDir = baseDir
		} else {
			outDir = filepath.Join(baseDir, "page", fmt.Sprintf("%d", page))
		}

		data := &theme.TemplateData{
			Site:   site,
			Page:   &content.Page{Title: title},
			Config: b.cfg,
			Extra: map[string]interface{}{
				"title":     title,
				"pages":     paginator.Pages,
				"paginator": paginator,
				"basePath":  basePath,
			},
		}

		if err := b.renderToFile(data, filepath.Join(outDir, "index.html"), tmpl); err != nil {
			return fmt.Errorf("render %s page %d: %w", title, page, err)
		}
	}
	return nil
}
