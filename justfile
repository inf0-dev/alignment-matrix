set unstable

out_dir := absolute_path("./_output")

# initialize directory if it does not exist
init_dir_ine(dir) := shell(f"mkdir -p {{ dir }}")

# run Go unit tests with coverage
go_test_unit:
    {{ init_dir_ine(out_dir) }}
    @ go test -tags unit ./... -coverprofile={{ out_dir }}/coverage.out

# run all tests
test: go_test_unit
