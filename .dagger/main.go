package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"dagger/slides/internal/dagger"
)

type Presentation struct {
	Path  string
	Title string
}

var presentations = []Presentation{
	{Path: "tlak", Title: "Thinking like a Kubernetes"},
}

type Slides struct{}

// Build builds all the Slidev presentations
// and merges them in a single directory.
//
// Use `dagger call build export --path _site --wipe`
// to generate the _site/ directory.
func (m *Slides) Build(
	ctx context.Context,
	// +defaultPath="/"
	// +ignore=["*", "!package.json", "!package-lock.json", "!tlak/"]
	source *dagger.Directory,
) (*dagger.Directory, error) {
	container := dag.Container().
		From("node:lts-alpine").
		WithFile("/src/package.json", source.File("package.json")).
		WithFile("/src/package-lock.json", source.File("package-lock.json")).
		WithWorkdir("/src").
		WithMountedCache("/root/.npm", dag.CacheVolume("npm")).
		WithExec([]string{"npm", "ci"})

	index, err := generateIndex(presentations)
	if err != nil {
		return nil, fmt.Errorf("failed to generate index.html: %w", err)
	}

	site := dag.Directory()
	for _, presentation := range presentations {
		deck := container.
			WithDirectory("/src/deck", source.Directory(presentation.Path)).
			WithWorkdir("/src/deck").
			WithExec([]string{"npx", "@slidev/cli", "build", "--base", "/" + presentation.Path}).
			Directory("dist").
			WithoutFile("_redirects") // Not supported by GitHub Pages
		site = site.WithDirectory(presentation.Path, deck)
	}

	return site.WithNewFile("index.html", index), nil
}

func generateIndex(presentations []Presentation) (string, error) {
	const indexTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Presentations</title>
</head>
<body>
    <h1>Presentations</h1>
    <ul>
{{- range .}}
        <li><a href="./{{.Path}}">{{.Title}}</a></li>
{{- end}}
    </ul>
</body>
</html>`

	tmpl, err := template.New("index").Parse(indexTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, presentations); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
