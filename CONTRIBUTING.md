# Contributing

## Generating the SDK

You can trigger a build with the GitHub action to generate the client. This will
automatically update the client to the latest version based on the Omicron commit hash
in the [`VERSION_OMICRON`](./VERSION_OMICRON) file.

Alternatively, if you wish to generate the client locally with your changes, update
the [`VERSION_OMICRON`](./VERSION_OMICRON) file with the git hash of the omicron branch
you wish to generate the SDK from, and run:

```bash
$ make all
```

## Backporting changes

The repository is organized with multiple release branches, each targeting a
specific release line. The release branches are named `rel/vX.Y` where `X.Y`
represents the release line version.

Pull requests should target the `main` branch and be backported to release
lines as necessary.

To backport a PR to the branch `rel/vX.Y` add the label
`backport/vX.Y` to the PR. Once merged, the backport automation will create a
new PR backporting the changes to the release branch. The backport label can
also be added after the PR is merged.

If a backport has merge conflicts, the conflicts are committed to the PR and
you can checkout the branch to fix them. Once the changes are clean, you can
merge the backport PR.

## Releasing a new SDK version

### Breaking change release

Releases that contain breaking changes require a new release branch.

1. Create a branch called `rel/vX.Y` from `main`.
2. Create a new label called `backport/vX.Y`.

Proceed with the steps below to complete the release.

### General release flow

1. Switch to the release branch you are targeting.
2. Make sure the following files have the new version you want to release.
   - [`VERSION`](./VERSION)
   - [`oxide/version.go`](./oxide/version.go)
3. Make sure you have run `make all` and pushed any changes. The release
   will fail if running `make all` causes any changes to the generated
   code.
4. Generate the changelog with `make changelog`.
5. Run `make tag` from your local `main` branch. This is just a command for making a git tag
   formatted correctly with the version.
6. Push the tag (the result of `make tag` gives instructions for this) to this repository.
7. Everything else is triggered from the tag push. Just make sure all the tests
   pass on the `main` branch before making and pushing a new tag.
