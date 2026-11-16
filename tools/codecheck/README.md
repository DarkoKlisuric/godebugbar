# codecheck

`codecheck` checks the Go code conventions that `golangci-lint` cannot enforce. It runs as part of `make lint`.

## What it checks

| Rule | Why |
| --- | --- |
| **No comments in Go code** | Code explains itself through clear names, small functions, and tests; `README.md` files explain package design and usage. Compiler directives (`//go:embed`, `//go:generate`, `//go:build`), `// Output:` blocks of runnable examples in `_test.go` files, and generated files are allowed. |
| **Every method has a pointer receiver** | Consistent receiver style across the codebase (`func (p *Profiler) ...`). |
| **A file declaring types declares one named after the file** | One concept per file in `snake_case` (`Profiler` in `profiler.go`, `GormPlugin` in `gorm_plugin.go`). |
| **Every Go package directory has a `README.md`** | Every package documents what it does and how to use it. |

## Running

```bash
go run ./tools/codecheck .
go test -race ./tools/codecheck/...
```
