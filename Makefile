.DEFAULT_GOAL := check

.PHONY: check format lint test eval build clean run

check: format lint test
	@echo "All checks passed."

format:
	@echo "Checking formatting..."
	@unformatted=$$(find . -name "*.go" -not -path "*/.*" -exec gofmt -l {} + 2>/dev/null); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files found:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

lint:
	@echo "Running static analysis..."
	@if find . -name "*.go" -not -path "*/.*" -print -quit | grep -q .; then \
		go vet ./...; \
	else \
		echo "No Go source files to lint yet."; \
	fi

test:
	@echo "Running unit tests..."
	@unit_pkgs=$$(go list ./... 2>/dev/null | grep -v "/eval$$" || true); \
	if [ -n "$$unit_pkgs" ]; then \
		go test -v $$unit_pkgs; \
	else \
		echo "No unit test packages found yet."; \
	fi

eval:
	@echo "Running regression evals (without live model calls)..."
	@go test -v -run "TestEval" ./eval/...

build:
	@echo "Building binary..."
	@if [ -d "cmd/adstudio" ]; then \
		go build -o bin/adstudio ./cmd/adstudio; \
	else \
		echo "cmd/adstudio not present yet."; \
	fi

clean:
	@rm -rf bin page.html brief.json trace.json

run:
	@if [ -n "$(IDEA)" ]; then \
		go run ./cmd/adstudio "$(IDEA)"; \
	else \
		go run ./cmd/adstudio "landing page for Strand, a fictional hair care subscription"; \
	fi

# Catch-all pattern: allows running `make "idea text"` directly
%:
	@go run ./cmd/adstudio "$@"
