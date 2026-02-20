default:
	just --list

# Start a local documentation server at http://localhost:6060/pkg/
documentation:
	go doc -all -u -http

# Build all packages
build:
	go build ./...

# Run tests with coverage reporting
test:
	go test ./... --cover -coverprofile=reports/coverage.out --covermode atomic --coverpkg=./...

# Show coverage report in browser
show-coverage-report:
	go tool cover -html=reports/coverage.out

# Alias to run tests and show coverage report
coverage-report: test show-coverage-report

# Runs code generation for all packages
generate:
	go generate ./...

# Run golangci-lint with auto-fix
lint:
	go tool golangci-lint run -v --fix

# Format all Go files
format:
	go fmt ./...

# Run all
checks: format lint test build

