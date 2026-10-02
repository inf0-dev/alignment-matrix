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

# run accessibility check on rendered demo (light + dark)
a11y: demo
    @ echo "--- light mode ---"
    @ pa11y {{ out_dir }}/demo.html || true
    @ sed 's/<html lang="en">/<html lang="en" data-theme="dark">/' {{ out_dir }}/demo.html > {{ out_dir }}/demo-dark.html
    @ echo "--- dark mode ---"
    @ pa11y {{ out_dir }}/demo-dark.html || true

# render demo HTML and open in browser
demo:
    {{ init_dir_ine(out_dir) }}
    @ go run ./app/cli render -p internal/pkg/renderer/testdata/demo.yaml -t record -o {{ out_dir }}/demo.html
    @ open {{ out_dir }}/demo.html

# update golden files for integration tests
update_golden:
    @ go test -tags=integration ./internal/pkg/exporter/ -update
    @ go test -tags=integration ./internal/pkg/renderer/ -update
