---
name: Release Checklist
about: Tracking issue for releasing a new version of the Go SDK.
title: Release for RXX
---

It's time to release a new version of the Go SDK. Use the checklist below to
perform the release.

## Release Checklist

- [ ] Ensure the `VERSION` file has the version to be released.
- [ ] Ensure the `.changelog/<VERSION>.toml` file has the changelog entries for
      the release.
- [ ] Update the examples and documentation to reference the new version.
- [ ] Fetch the latest tags from remote and generate the changelog using `make
      changelog`. Update the generated file with the release date and associated
      Oxide API version.
- [ ] Create the release tag using `make tag`.
- [ ] Update the GitHub release description with the release content generated
      from `make changelog`.
- [ ] Create and push a release branch from the commit of the release tag (e.g.,
      `X.Y` for release `vX.Y.Z`).
- [ ] Update the `VERSION` file with the next development version and run `make
      generate` to update generated files.
- [ ] Create a new `.changelog/<VERSION>.toml` file with the next development
      version.
