# ADR-0001: Value receivers for App

**Date:** 2026-02-28

## Context and Problem Statement

tdocker uses the [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI framework, which is modelled after the Elm architecture. The central type, `App`, holds the full application state and implements the `tea.Model` interface (`Init`, `Update`, `View`).

In Go, methods can be defined with either value receivers (`func (m App)`) or pointer receivers (`func (m *App)`). For a large struct with many fields, pointer receivers are often the default choice to avoid copying. The question was whether `App` should follow that convention.

## Considered Options

* **Value receivers** - `Update` receives a copy of `App`, modifies it, and returns the modified copy alongside a command. All state-mutating helpers (e.g. `rebuildTable`, `closeLogs`) return a new `App`.
* **Pointer receivers** - `Update` and helpers mutate `App` in place; `tea.NewProgram` is called with `&App{}`.

## Decision Outcome and Drivers

Chosen option: **Value receivers**, because

* Bubbletea is designed around the Elm architecture, where `Update` is a pure function: `(model, message) → (model, command)`. Value receivers make this contract explicit in the type system - the caller always gets a new model back, never a silently mutated one.
* The framework copies the model on every message dispatch regardless, so switching to pointer receivers would not eliminate copying - it would just move where it happens while adding aliasing risk.
* Value receivers make the test helpers (`update`, `confirming`, `statsPanel`) straightforward: they take and return `App` by value with no pointer indirection or type assertions against `*App`.
* Pointer receivers would require passing `&App{}` to `tea.NewProgram` and changing all `result.(App)` type assertions in tests to `result.(*App)`, for no practical gain.

## People
- @pivovarit
