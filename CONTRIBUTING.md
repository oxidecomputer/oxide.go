# Contributing

## Generating the SDK

You can trigger a build with the GitHub action to generate the client. This will
automatically update the client to the latest version based on the Omicron commit hash
in the [`VERSION_OMICRON`](./VERSION_OMICRON) file.

Alternatively, if you wish to generate the client locally, run:

```bash
$ make all
```

## Releasing a new SDK version

1. Make sure the [`VERSION.txt`](./VERSION.txt) file has the new version you want to release.
2. Make sure you have run `make all` and pushed any changes. The release
   will fail if running `make all` causes any changes to the generated
   code.
3. Run `make tag` from your local `main` branch. This is just a command for making a git tag
   formatted correctly with the version.
4. Push the tag (the result of `make tag` gives instructions for this) to this repository.
5. Everything else is triggered from the tag push. Just make sure all the tests
   pass on the `main` branch before making and pushing a new tag.
