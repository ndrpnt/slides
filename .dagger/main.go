package main

import (
	"context"
	"fmt"

	"dagger/slides/internal/dagger"
)

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

	var redirects string
	var index string
	site := dag.Directory()

	for _, path := range []string{"tlak"} {
		deck := container.
			WithDirectory("/src/deck", source.Directory(path)).
			WithWorkdir("/src/deck").
			WithExec([]string{"npx", "@slidev/cli", "build", "--base", "/" + path}).
			Directory("dist")
		content, err := deck.File("_redirects").Contents(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read _redirects: %w", err)
		}
		redirects += content
		index += "        <li><a href=\"./" + path + "\">Thinking like a Kubernetes</a></li>\n"
		site = site.WithDirectory(path, deck.WithoutFile("_redirects"))
	}

	site = site.
		WithNewFile("_redirects", redirects).
		WithNewFile("index.html", "<!DOCTYPE html>\n<html>\n<head>\n    <title>Presentations</title>\n</head>\n<body>\n    <h1>Presentations</h1>\n    <ul>\n"+index+"    </ul>\n</body>\n</html>")

	return site, nil
}
