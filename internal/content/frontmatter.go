package content

import (
	"bufio"
	"bytes"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const fmDelim = "---"

func ParseFrontMatter(raw []byte) (*Page, []byte, error) {
	if !bytes.HasPrefix(raw, []byte(fmDelim)) {
		return &Page{RawContent: string(raw)}, raw, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Scan() // skip first ---

	var yBuf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == fmDelim {
			break
		}
		yBuf.WriteString(line + "\n")
	}

	bodyStart := len(fmDelim) + 1 + yBuf.Len() + len(fmDelim)
	if bodyStart > len(raw) {
		bodyStart = len(raw)
	}
	body := bytes.TrimSpace(raw[bodyStart:])

	page := &Page{RawContent: string(body)}
	if err := parseFM(yBuf.String(), page); err != nil {
		return page, body, err
	}
	return page, body, nil
}

type fmRaw struct {
	Title, Date, LastMod, Slug, URL, Description, Summary string
	Draft                                                  bool
	Categories, Tags                                       []string
	Weight                                                 int
	TOC, Math                                              *bool
}

func parseFM(data string, page *Page) error {
	var raw fmRaw
	if err := yaml.Unmarshal([]byte(data), &raw); err != nil {
		return err
	}
	page.Title = raw.Title
	page.Draft = raw.Draft
	page.Categories = raw.Categories
	page.Tags = raw.Tags
	page.Slug = raw.Slug
	page.URL = raw.URL
	page.Weight = raw.Weight
	page.Description = raw.Description
	page.Summary = raw.Summary
	if raw.TOC != nil {
		page.TOC = *raw.TOC
	}
	if raw.Math != nil {
		page.Math = *raw.Math
	}
	if raw.Date != "" {
		t, err := parseDate(raw.Date)
		if err != nil {
			return err
		}
		page.Date = t
	}
	if raw.LastMod != "" {
		t, err := parseDate(raw.LastMod)
		if err != nil {
			return err
		}
		page.LastMod = t
	}
	return nil
}

func parseDate(s string) (time.Time, error) {
	for _, f := range []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
	} {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}
