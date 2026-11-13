# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-13

First tagged release. The library, CLI, REST server, and optional GUI match the rust-sudoku 1.0.1 behavior.

### Security

- Require Go 1.26.10 or newer. Go 1.25 and older are end of life and do not include the October 2026 toolchain fixes.
- Require `golang.org/x/net` v0.60.0 for the Fyne GUI.

### Changed

- CLI and server report version `1.0.0` unless the build injects another value with `-ldflags`.
- Document that `Grid` accepts sizes up to 25×25. The GUI presets stay 4×4, 6×6, and 9×9.
- The Docker builder is `golang:1.26.10-alpine`.

## [Unreleased]

[1.0.0]: https://github.com/RumenDamyanov/go-sudoku/releases/tag/v1.0.0
