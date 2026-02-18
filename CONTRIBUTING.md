# Contributing

## Generating the SDK

You can trigger a build with the GitHub action to generate the client. This will automatically
update the client to the latest version based on the Omicron commit hash in the
[`VERSION_OMICRON`](./VERSION_OMICRON) file.

Alternatively, if you wish to generate the client locally with your changes, update the
[`VERSION_OMICRON`](./VERSION_OMICRON) file with the git hash of the omicron branch you wish to
generate the SDK from, and run:

```bash
$ make all
```

## Releasing a new SDK version

The release process requires you to sign the release tag. Before starting the
release, ensure you have Git configured with a signing key.

Git 2.34 or later can use SSH keys for signing, which can be easier to
configure then GPG. Here is an example of what to add to your `.gitconfig`
file.

```
[user]
signingkey = <PATH TO SSH PUBLIC KEY>

[gpg]
format = ssh
```

- [ ] Create a release branch
  ```
  git checkout main
  git pull origin main
  git checkout -b release-vX.Y.Z
  ```
- [ ] Update the [`VERSION`](./VERSION) and [`VERSION_OMICRON`](./VERSION_OMICRON) files with the new version you want to release.
- [ ] Generate and lint files.
  ```
  make all
  ```
- [ ] Ensure the `.changelog/vX.Y.Z.toml` file has the changelog entries for the release and generate changelog.
  ```
  make changelog
  ```
- [ ] Update the generated file with the release date and associated Oxide API version.
  ```diff
  - # vX.Y.Z
  + # vX.Y.Z (Year/Month/Day)
  +
  + Generated from Oxide API version [API VERSION](https://github.com/oxidecomputer/omicron/blob/<OMICRON TAG>/openapi/nexus/nexus-<API VERSION>.json)
  ```
- [ ] Commit and push updated files.
  ```
  git add -A .
  git commit -m 'release vX.Y.Z'
  git push origin release-vX.Y.Z
  ```
- [ ] Open a PR to update `main`.
- [ ] Ensure tests are passing in `main` after merge.
- [ ] Run `make tag` from your local `main` branch.
  ```
  git checkout main
  git pull origin main
  make tag
  ```
- [ ] Push the tag to this repository.
  ```
  git push origin vX.Y.Z
  ```
- [ ] Update the GitHub [release](https://github.com/oxidecomputer/oxide.go/releases) description with the release content generated from `make changelog`.
- [ ] Create and push a release branch from the commit of the release tag.
  ```
  git checkout vX.Y.Z
  git checkout -b X.Y
  git push origin X.Y
  ```
- [ ] Create a new branch to prepare the repository for the next version.
- [ ] Update the `VERSION` file with the next development version and run `make generate` to update generated files.
- [ ] Create a new `.changelog/<VERSION>.toml` file with the next development version.
- [ ] Push changes and open PR.
