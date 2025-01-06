---
name: Release checklist
about: Steps to take when releasing a new version (only for Oxide release team).
labels: release

---

## Release checklist
<!-- 
 Please follow all of these steps in the order below.
 After completing each task put an `x` in the corresponding box,
 and paste the link to the relevant PR.
-->
- [ ] Make sure the [VERSION](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION) and [oxide/version.go](https://github.com/oxidecomputer/oxide.go/blob/main/oxide/version.go) files have the new version you want to release.
- [ ] Make sure the changelog file in the `.changelog/` directory is set to the new version you want to release (e.g., replace `+dev` metadata).
- [ ] Make sure all examples and docs reference the new version.
- [ ] Generate changelog by running `make changelog` and add date of the release to the title.
- [ ] Release the new version by running `make tag`.
- [ ] Update GitHub release description with release notes generated from `make changelog`.
- [ ] Create a release branch from the commit of the release tag.
- [ ] Bump the version in [VERSION](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION) and [oxide/version.go](https://github.com/oxidecomputer/oxide.go/blob/main/oxide/version.go). If the `version` within [`nexus.json`](https://github.com/oxidecomputer/omicron/blob/main/openapi/nexus.json) hasn't been updated to the next release, then use `+dev` for the version metadata.
- [ ] Create a new file for the next release in [.changelog/](https://github.com/oxidecomputer/oxide.go/blob/main/.changelog/). If the `version` within [`nexus.json`](https://github.com/oxidecomputer/omicron/blob/main/openapi/nexus.json) hasn't been updated to the next release, then use `+dev` for the version metadata.
