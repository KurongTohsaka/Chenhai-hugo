package content_test

import (
	"strings"
	"testing"
	"time"

	"github.com/KurongTohsaka/chenhai-hugo/internal/content"
)

func TestParseFrontMatter_Full(t *testing.T) {
	raw := `---
title: My Post
date: 2024-01-15
draft: true
categories:
  - go
  - hugo
tags:
  - blog
  - static
slug: my-post
toc: false
math: true
weight: 10
description: A test post
summary: Short summary
lastmod: 2024-01-20
---
# Hello

This is the body.
`
	page, body, err := content.ParseFrontMatter([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.Title != "My Post" {
		t.Errorf("Title = %q, want %q", page.Title, "My Post")
	}
	if !page.Date.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Date = %v, want 2024-01-15", page.Date)
	}
	if !page.LastMod.Equal(time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("LastMod = %v, want 2024-01-20", page.LastMod)
	}
	if !page.Draft {
		t.Error("Draft should be true")
	}
	if len(page.Categories) != 2 || page.Categories[0] != "go" || page.Categories[1] != "hugo" {
		t.Errorf("Categories = %v, want [go hugo]", page.Categories)
	}
	if len(page.Tags) != 2 || page.Tags[0] != "blog" || page.Tags[1] != "static" {
		t.Errorf("Tags = %v, want [blog static]", page.Tags)
	}
	if page.Slug != "my-post" {
		t.Errorf("Slug = %q, want %q", page.Slug, "my-post")
	}
	if page.TOC == nil || *page.TOC {
		t.Error("TOC should be false")
	}
	if page.Math == nil || !*page.Math {
		t.Error("Math should be true")
	}
	if page.Weight != 10 {
		t.Errorf("Weight = %d, want 10", page.Weight)
	}
	if page.Description != "A test post" {
		t.Errorf("Description = %q, want %q", page.Description, "A test post")
	}
	if page.Summary != "Short summary" {
		t.Errorf("Summary = %q, want %q", page.Summary, "Short summary")
	}
	if string(body) != "# Hello\n\nThis is the body." {
		t.Errorf("body = %q, want %q", string(body), "# Hello\n\nThis is the body.")
	}
}

func TestParseFrontMatter_NoFrontMatter(t *testing.T) {
	raw := `Just some plain text without front matter.`
	page, body, err := content.ParseFrontMatter([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Title != "" {
		t.Errorf("Title should be empty, got %q", page.Title)
	}
	if string(body) != raw {
		t.Errorf("body = %q, want %q", string(body), raw)
	}
}

func TestParseFrontMatter_Minimal(t *testing.T) {
	raw := `---
title: Hello
date: "2024-06-01"
---
Some content.
`
	page, body, err := content.ParseFrontMatter([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Title != "Hello" {
		t.Errorf("Title = %q, want %q", page.Title, "Hello")
	}
	if !page.Date.Equal(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Date = %v, want 2024-06-01", page.Date)
	}
	// Defaults should not be overridden
	if page.Draft {
		t.Error("Draft should be false")
	}
	if page.TOC != nil {
		t.Error("TOC should be nil")
	}
	if page.Weight != 0 {
		t.Errorf("Weight should be 0, got %d", page.Weight)
	}
	if string(body) != "Some content." {
		t.Errorf("body = %q, want %q", string(body), "Some content.")
	}
}

func TestParseFrontMatter_InvalidDate(t *testing.T) {
	raw := `---
title: Test
date: "not-a-date"
---
Body.
`
	_, _, err := content.ParseFrontMatter([]byte(raw))
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
	if !strings.Contains(err.Error(), "invalid date") {
		t.Errorf("error should contain 'invalid date', got: %v", err)
	}
}

func TestParseFrontMatter_Unclosed(t *testing.T) {
	raw := `---
title: Broken
Body.
`
	_, _, err := content.ParseFrontMatter([]byte(raw))
	if err == nil {
		t.Fatal("expected error for unclosed front matter, got nil")
	}
	if !strings.Contains(err.Error(), "unclosed front matter") {
		t.Errorf("error should contain 'unclosed front matter', got: %v", err)
	}
}

func TestParseFrontMatter_CoverAndPinned(t *testing.T) {
	raw := `---
title: Featured Post
date: 2024-06-01
cover: https://example.com/cover.jpg
pinned: true
---
Body content.
`
	page, body, err := content.ParseFrontMatter([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.Cover != "https://example.com/cover.jpg" {
		t.Errorf("Cover = %q, want %q", page.Cover, "https://example.com/cover.jpg")
	}
	if !page.Pinned {
		t.Error("Pinned should be true")
	}
	if string(body) != "Body content." {
		t.Errorf("body = %q, want %q", string(body), "Body content.")
	}
}

func TestParseFrontMatter_CoverAndPinned_Default(t *testing.T) {
	raw := `---
title: Normal Post
date: 2024-06-01
---
Body.
`
	page, _, err := content.ParseFrontMatter([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.Cover != "" {
		t.Errorf("Cover should be empty by default, got %q", page.Cover)
	}
	if page.Pinned {
		t.Error("Pinned should be false by default")
	}
}
