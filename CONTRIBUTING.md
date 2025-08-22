# Contributing

Thanks for your interest in improving `go-sudoku`.

## Quick Start

1. Fork & clone the repo.
2. Ensure you have Go 1.22+.
3. Run tests: `go test ./...` (should be green).
4. Make focused changes (small, single-topic commits).
5. Add/adjust tests when you change behavior.
6. Run `make build` (ensures version ldflags still work) and `go test ./...` again.
7. Open a PR with a clear title & short description.

## Guidelines

- Keep the core library stdlib-only. GUI dependencies stay behind the `gui` build tag.
- Preserve public API stability; propose breaking changes in an issue first.
- Favor clarity over micro-optimizations unless a bottleneck is demonstrated.
- Keep README examples in sync; update or add an example under `examples/` when adding notable features.
- New generator / solver logic must retain uniqueness guarantees (add a test).
- Use `SetRandSeed` in tests/examples if deterministic behavior is needed.

## Commit Messages

Use concise, present tense: `add grid attempts retry`, `fix uniqueness check for 6x6`, `docs: clarify /health alias`.

## Reporting Issues

Include: Go version, OS, minimal reproducible snippet / puzzle string, expected vs actual behavior.

## Security

Do not open public issues for vulnerabilities. Follow the steps in `SECURITY.md`.

## Code of Conduct

Participation implies agreement with `CODE_OF_CONDUCT.md`.

## License

By contributing you agree your changes are MIT licensed as per `LICENSE.md`.

Happy puzzling! 🧩
