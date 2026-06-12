package config

type Config struct {
	Title       string          `yaml:"title"`
	Subtitle    string          `yaml:"subtitle"`
	Description string          `yaml:"description"`
	BaseURL     string          `yaml:"baseURL"`
	Language    string          `yaml:"language"`
	Copyright   string          `yaml:"copyright"`
	Author      Author          `yaml:"author"`
	Theme       string          `yaml:"theme"`
	ThemeConfig ThemeConfig     `yaml:"themeConfig"`
	Menu        []MenuItem      `yaml:"menu"`
	Markup      Markup          `yaml:"markup"`
	Social      Social          `yaml:"social"`
	SEO         SEO             `yaml:"seo"`
	ImageHost       ImageHostConfig     `yaml:"imageHost"`
	TagDescriptions map[string]string  `yaml:"tagDescriptions"`
}

type Author struct {
	Name   string `yaml:"name"`
	Avatar string `yaml:"avatar"`
	Bio    string `yaml:"bio"`
}

type ThemeConfig struct {
	ColorMode    string                 `yaml:"colorMode"`
	ShowToc      bool                   `yaml:"showToc"`
	TocFloat     bool                   `yaml:"tocFloat"`
	CodeTheme    string                 `yaml:"codeTheme"`
	DateFormat   string                 `yaml:"dateFormat"`
	PostsPerPage int                    `yaml:"postsPerPage"`
	Params       map[string]interface{} `yaml:",inline"`
}

type MenuItem struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	Icon string `yaml:"icon"`
}

type Markup struct {
	Highlight HighlightConfig `yaml:"highlight"`
	Math      MathConfig      `yaml:"math"`
	Mermaid   bool            `yaml:"mermaid"`
	TOC       TOCConfig       `yaml:"toc"`
}

type HighlightConfig struct {
	Style        string `yaml:"style"`
	LineNumbers  bool   `yaml:"lineNumbers"`
	ShowFilename bool   `yaml:"showFilename"`
}

type MathConfig struct{ Engine string `yaml:"engine"` }

type TOCConfig struct {
	MinDepth int `yaml:"minDepth"`
	MaxDepth int `yaml:"maxDepth"`
}

type Social struct {
	GitHub  string `yaml:"github"`
	Email   string `yaml:"email"`
	Twitter string `yaml:"twitter"`
}

type SEO struct {
	GoogleAnalytics string `yaml:"googleAnalytics"`
	EnableRobotsTXT bool   `yaml:"enableRobotsTXT"`
	EnableSitemap   bool   `yaml:"enableSitemap"`
}

type ImageHostConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	Repo     string `yaml:"repo"`
	Branch   string `yaml:"branch"`
	BasePath string `yaml:"basePath"`
	Token    string `yaml:"token"`
	Mode     string `yaml:"mode"`
	BaseURL  string `yaml:"baseURL"`
}

func DefaultConfig() *Config {
	return &Config{
		Language: "zh-CN",
		Theme:    "zhenhai",
		ThemeConfig: ThemeConfig{
			ColorMode: "auto", ShowToc: true, TocFloat: true,
			CodeTheme: "github-dark", DateFormat: "2006-01-02", PostsPerPage: 10,
			Params:    map[string]interface{}{},
		},
		Markup: Markup{
			Highlight: HighlightConfig{Style: "github-dark", LineNumbers: true, ShowFilename: true},
			Math:      MathConfig{Engine: "katex"},
			Mermaid:   true,
			TOC:       TOCConfig{MinDepth: 2, MaxDepth: 4},
		},
		SEO: SEO{EnableRobotsTXT: true, EnableSitemap: true},
		ImageHost: ImageHostConfig{
			Provider: "github",
			Branch:   "main",
			BasePath: "images/",
			Mode:     "auto",
		},
	}
}
