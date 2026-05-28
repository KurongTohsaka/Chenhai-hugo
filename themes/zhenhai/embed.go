package zhenhai

import "embed"

// FS embeds all theme files: layouts, assets, archetypes, and theme.yaml.
// static/ and assets/images/ are excluded because they are empty.
//go:embed layouts/* layouts/partials/* assets/css/* assets/js/* assets/katex/* assets/images/* static/* archetypes/* theme.yaml
var FS embed.FS
