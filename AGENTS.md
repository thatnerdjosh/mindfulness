# Resume Prompt

You are a coding assistant resuming work on the Mindfulness (mt) project. The user wants a slow, buffer-friendly walkthrough and iterative changes.

## Project Summary

- DDD-style Go project to journal daily reflections on the Five Mindfulness Trainings.
- Layers: domain (precepts + entry rules), application service, in-memory repository, CLI interface, tests.
- CLI supports: `mt journal add`, `mt journal guided`, `mt journal latest`, `mt journal list`, `mt version`.
- Guided flow prompts: date, mood, note, each precept, then optional summary + confirmation.
- High test coverage across layers; CLI ~95.8% in last run.

## Current Files of Interest

- Domain: `internal/domain/journal/precepts.go`, `internal/domain/journal/entry.go`, `internal/domain/journal/repository.go`
- Application: `internal/application/journal/service.go`
- Infrastructure: `internal/infrastructure/persistence/memory/journal_repository.go`
- CLI: `internal/interfaces/cli/app.go`
- Entrypoint: `cmd/mt/main.go`
- Code reader prompt: `prompts/code-reader.md`
- Reports: `vibe-code/report-2026-01-12-initial.md`, `vibe-code/report-2026-01-12-updated.md`

## User Preferences

- Very small output chunks; user cannot scroll easily.
- Start high-level and show only the most important parts first.
- Provide one snippet at a time; pause and wait for "." to continue when a read-through is requested.
- Only do snippet read-throughs when explicitly requested.
- Avoid findings/critique unless asked; focus on understanding.
- For any redesign work, propose a plan and check in before coding.

## Suggested Resume Prompt (verbatim)

"Please resume the Mindfulness (mt) project. Use the code-reader prompt in `prompts/code-reader.md`, and continue the guided walkthrough or next tasks in small buffer-friendly steps. Ask me for confirmation before making major changes."
