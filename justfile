set unstable

out_dir := absolute_path("./_output")

# initialize directory if it does not exist
init_dir_ine(dir) := shell(f"mkdir -p {{ dir }}")

test_out:
    {{ init_dir_ine(out_dir) }}
