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
- [ ] Make sure the following files have the new version you want to release.
  - [ ] [`VERSION`](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION)
  - [ ] [`oxide/version.go`](https://github.com/oxidecomputer/oxide.go/blob/main/oxide/version.go)
- [ ] Make sure all examples and docs reference the new version.
- [ ] Generate changelog by running `make changelog` and add date of the release to the title.
- [ ] Release the new version by running `make tag`.
