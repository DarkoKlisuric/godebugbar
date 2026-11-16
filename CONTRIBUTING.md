# Contributing

## In short

1. Branch off `main`: `feat/<short-name>`, `fix/<short-name>`, or `chore/<short-name>`.
2. Commit with [Conventional Commits](https://www.conventionalcommits.org/) messages.
3. Push the branch and open a pull request into `main`.
4. CI must pass: the repository checks (`ci.yml`) and the commit message check (`commitlint.yml`). Both run on pull requests.
5. Merge with **Create a merge commit**, then delete the branch.
6. Releases happen automatically: `release-please` opens a release pull request and merges it when checks pass.

## Branches

- `main` is protected: nobody pushes to it directly; every change arrives through a pull request.
- A branch lives for one logical change. If `main` moved on, update the branch from `main` before merging.
- Names: `feat/<short-name>` for a feature, `fix/<short-name>` for a bug fix, `chore/<short-name>` for everything else.

## Commit messages

```
<type>: <what the commit does, imperative>

<optional body: why>

<optional footer: BREAKING CHANGE: what breaks and how to migrate>
```

| Type | When | Release |
| --- | --- | --- |
| `feat` | something new a user can see or use | minor (`v1.2.0` → `v1.3.0`) |
| `fix` | a bug fix | patch (`v1.2.0` → `v1.2.1`) |
| `feat!` / `fix!`, or a `BREAKING CHANGE:` footer | breaking API change | major (`v1.2.0` → `v2.0.0`) |
| `docs`, `test`, `refactor`, `style`, `perf`, `build`, `ci`, `chore` | everything else | none |

While the repository is below `v1.0.0`, a breaking change raises the minor version (`v0.1.0` → `v0.2.0`) instead of the major one.

Every commit on `main` counts because pull requests are merged with a merge commit, not squashed or rebased. Keep one logical change per commit and do not include AI co-author attribution lines.

## Pull requests

- Target `main`. The title follows the same format as a commit message.
- Merge only when every check is green, with **Create a merge commit**. Squash merging and rebase merging are turned off.
- Delete the branch after merging.

## Releases

Nobody raises versions or creates tags by hand.

1. After each merge into `main`, [release-please](https://github.com/googleapis/release-please) opens or updates one release pull request with the next version and `CHANGELOG.md` entries.
2. The release pull request changes only `CHANGELOG.md` files and version numbers, so `ci.yml` jobs skip it and only the commit message check runs. The release workflow enables auto-merge for that pull request so it merges into `main` automatically once required checks pass.
3. Merging the release pull request tags the commit `vX.Y.Z` for the core and `adapter/<name>/vX.Y.Z` for each adapter module, and creates the GitHub Releases. All modules share one version, so `go get github.com/DarkoKlisuric/godebugbar@vX.Y.Z` and `go get github.com/DarkoKlisuric/godebugbar/adapter/gin@vX.Y.Z` work right away.
4. Never move or delete a pushed tag. A mistake is fixed by the next release.
