.PHONY: dev build check clean fmt vet

# Live dev server on :8080. Templates, CSS and docs are read from disk,
# so a browser refresh picks up edits -- no rebuild needed.
dev:
	go run . serve -addr :8080

# Render the static site into ./public (what GitHub Pages publishes).
build:
	go run . build

# Validate content without writing anything.
check:
	go run . check

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -rf public
