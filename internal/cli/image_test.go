package cli

import "testing"

func TestPostToImgDir(t *testing.T) {
	cases := []struct{ post, want string }{
		{"posts/CS224N/lesson_5.md", "img/CS224N/lesson_5"},
		{"posts/hello.md", "img/hello"},
		{"about/index.md", "img/about"},
		{"posts/DeepDive/README.md", "img/DeepDive/README"},
	}
	for _, c := range cases {
		if got := postToImgDir(c.post); got != c.want {
			t.Errorf("postToImgDir(%q) = %q, want %q", c.post, got, c.want)
		}
	}
}
