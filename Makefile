default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

release: 
	@echo "Releasing new version..."
	@echo "Please make sure to update the version in CHANGELOG.md before running this command."
	@read -p "Have you updated the version in CHANGELOG.md? (y/n) " -n 1 -r; echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		VERSION=$$(grep -Eo '## [0-9]+\.[0-9]+\.[0-9]+' CHANGELOG.md | head -1 | awk '{print $$2}'); \
		git tag -a "v$$VERSION" -m "Release version $$VERSION"; \
		git push origin "v$$VERSION"; \
		gh release create "v$$VERSION" --title "Release v$$VERSION" --notes-file CHANGELOG.md; \
		echo "Released version $$VERSION successfully."; \
	else \
		echo "Aborting release. Please update the version in CHANGELOG.md before running this command."; \
	fi

.PHONY: fmt lint test testacc build install generate
