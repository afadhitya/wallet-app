## 1. CI Workflow Update

- [x] 1.1 Add `if: github.ref == 'refs/heads/main'` to the `Upload coverage reports to Codecov` step in `.github/workflows/go-quality.yml`

## 2. Verification

- [x] 2.1 Verify the workflow YAML is valid (e.g., via `yamllint` or GitHub Actions schema)
- [x] 2.2 Confirm CI badge URL in `README.md` points to the correct workflow file
