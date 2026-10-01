set unstable

out_dir := absolute_path("./_output")
cov_dir := out_dir / "coverage"

# initialize directory if it does not exist
init_dir_ine(dir) := shell(f"mkdir -p {{ dir }}")

# run Go unit tests with coverage
go_test_unit:
    {{ init_dir_ine(cov_dir) }}
    @ go test -tags unit ./... -coverprofile={{ cov_dir }}/unit.out

# run Go integration tests with coverage
go_test_integration:
    {{ init_dir_ine(cov_dir) }}
    @ go test -tags integration ./... -coverprofile={{ cov_dir }}/integration.out

# run all tests and merge coverage
test: go_test_unit go_test_integration test_coverage

# merge unit + integration coverage into a single report
test_coverage:
    @ go tool gocovmerge {{ cov_dir }}/unit.out {{ cov_dir }}/integration.out > {{ cov_dir }}/merged.out
    @ echo "--- merged coverage ---"
    @ go tool cover -func={{ cov_dir }}/merged.out | grep total

lint:
    @ golangci-lint run --timeout 5m

fmt:
    @ go fmt ./...
