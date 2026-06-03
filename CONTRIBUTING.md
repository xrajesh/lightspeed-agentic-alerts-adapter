# Contributing

## Getting started

1. Fork and clone the repository
2. Install prerequisites: Go 1.26+, [golangci-lint](https://golangci-lint.run/)
3. Run `make test` and `make lint` to verify your setup

## Development workflow

### Proposing changes

This project uses the [OpenSpec](https://github.com/Fission-AI/OpenSpec) framework to manage specs and changes. Before starting implementation, propose your change through OpenSpec:

1. Run `/openspec-propose` (or `/opsx:propose`) to create a proposal with design, specs, and tasks
2. Get the proposal reviewed and approved
3. Implement the change following the generated tasks
4. Run `/openspec-verify-change` to verify the implementation matches the spec
5. Archive the change with `/openspec-archive-change` once merged

Specs live in [`openspec/specs/`](openspec/specs/) and changes are tracked in [`openspec/changes/`](openspec/changes/).

### Code changes

1. Create a feature branch from `main`
2. Make your changes
3. Ensure all checks pass:
   ```sh
   make fmt
   make vet
   make lint
   make test
   ```
4. Open a pull request

### Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add alertmanager retry logic
fix: guard against nil alert fingerprint
test: cover AlreadyExists path
docs: update architecture diagram
refactor: return (bool, error) from CreateProposal
```

## Code review

Pull requests require approval from a reviewer listed in the [OWNERS](OWNERS) file.

## License

By contributing, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).
