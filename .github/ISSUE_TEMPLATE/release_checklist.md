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

- [ ] Make sure the [VERSION](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION) file has
      the new version you want to release.
- [ ] Make sure the changelog file in the `.changelog/` directory is set to the new version you want
      to release.
- [ ] Make sure all examples and docs reference the new version.
- [ ] Make sure you've pulled the latest tag on main, and generate changelog by running
      `make changelog`. Add the date of the release to the title, and update associated Oxide API
      version.
- [ ] Release the new version by running `make tag`.
- [ ] Update GitHub release description with release notes generated from `make changelog`.
- [ ] Create a release branch from the commit of the release tag.
- [ ] Bump the version in [VERSION](https://github.com/oxidecomputer/oxide.go/blob/main/VERSION) and
      run `make generate` to update the generated files.
- [ ] Create a new file for the next release in
      [.changelog/](https://github.com/oxidecomputer/oxide.go/blob/main/.changelog/).
