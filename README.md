# debtCLI

CLI tool that scans a Git repository for technical debt markers, scores them by age and churn, and tracks repo health over time.

## Install

```
go install github.com/ashwinpatri/debtCLI@latest
```

Or build from source:

```
make build
./bin/debt
```

## Usage

```
debt scan [path]      scan repo, store snapshot, print report
debt history [path]   show health score over time
debt show <file>      show debt items for a specific file
debt init [path]      create .debt.toml in current repo
```

**Flags**

```
--format table|json   output format (default: table)
```

## Configuration (.debt.toml)

Optional. Searched by walking up from the target path. Defaults are used silently if not found.

```toml
[tags]
TODO     = 2.0
FIXME    = 3.0
HACK     = 2.5
PERF     = 2.0
SECURITY = 4.0
NOTE     = 1.0

[ignore]
paths      = ["vendor/", "node_modules/", ".git/", "dist/", "build/"]
extensions = [".pb.go", ".gen.go", ".min.js", ".lock", ".sum"]
```

Tag severity must be in range `(0, 10]`. Ignore paths must be relative and cannot contain `..`.

## Scoring

```
score  = severity × age_mult × churn_mult

age_mult   = 1 + min(age_days / 180, 2.0)   — caps at 3× (~1 year)
churn_mult = 1 + min(churn / 50, 1.0)        — caps at 2× (50 commits)

health = max(0, 100 − (sum(scores) / 200 × 100))
```

100 = clean. 0 = on fire.

## Development

```
make build    # build to bin/debt
make test     # go test ./... -race -count=1
make lint     # golangci-lint run ./...
make install  # go install
```

Requires Go 1.22+ and git on PATH.

## License

MIT
