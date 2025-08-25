# Guidelines for contributing

The current version is written on Linux Ubuntu 24.04 LTS and works only on it. Feel free to make
you own versions for other operating systems. Use `Makefile` file to run build and run the project.

## How to contribute

You can open an issue in GitHub or check out the [plans.md](plans.md) file for new features to be added or bugs to be destroyed. Write tests.

## Coding style

- Follow Golang's best practices (I use gofmt)
- Keep functions small and readable
- Write comments for exported functions and types.

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- feat: new feature
- fix: bug fix
- refactor: cleanup without behavior change
- docs: documentation only

Ex:

```
feat(cli): add json output mode
fix(parser): handle empty comment lines correctly
```

## Pull Request Process

- Make sure go test ./... passes.
- Update docs (README, TODO list) if your change adds/removes behavior.
- Keep PRs focused (one feature/fix at a time).
- Submit PR to main branch.

Happy Hacking!

