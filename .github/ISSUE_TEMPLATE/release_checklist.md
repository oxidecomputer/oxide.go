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
- [ ] Make sure all examples and docs reference the new version.
- [ ] Generate changelog by running `make changelog` and add date of the release to the title.
- [ ] Release the new version by running `make tag`.
- [ ] Bump the version in [VERSION](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION) and [oxide/version.go](https://github.com/oxidecomputer/oxide.go/blob/main/oxide/version.go)
- [ ] Create a new file for the next release in [.changelog/](https://github.com/oxidecomputer/oxide.go/blob/main/.changelog/)
