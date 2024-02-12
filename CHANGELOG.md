# v0.1.0-beta3 (2023/Feb/13)

### Breaking changes

- **Go version:** Minimum required Go version has been updated to 1.21. [#179](https://github.com/oxidecomputer/oxide.go/pull/179)
- **NewClient API change:** The `NewClient` function has been updated to no longer require a user agent parameter. [#180](https://github.com/oxidecomputer/oxide.go/pull/180)
- **NewClientFromEnv removal:** The `NewClientFromEnv` function has been removed. Users should use `NewClient` instead. [#180](https://github.com/oxidecomputer/oxide.go/pull/180)
- **Method renames:** Several methods have had slight name changes to better reflect their functionality. [#182](https://github.com/oxidecomputer/oxide.go/pull/182)
- **Types:** Several types have added fields and/or renames. [#182](https://github.com/oxidecomputer/oxide.go/pull/182), [#185](https://github.com/oxidecomputer/oxide.go/pull/185), [#190](https://github.com/oxidecomputer/oxide.go/pull/190)

### New features

- **New instance APIs:** Live attach and detach of external IPs to an instance. [#182](https://github.com/oxidecomputer/oxide.go/pull/182)
- **New IP pool APIs:** Several silo IP pool maintenance endpoints. [#182](https://github.com/oxidecomputer/oxide.go/pull/182), [#187](https://github.com/oxidecomputer/oxide.go/pull/187)
- **New SSH keys APIs:** Endpoint to list SSH keys that were added to an instance on create. [#185](https://github.com/oxidecomputer/oxide.go/pull/185)
- **New networking APIs:** Enable, disable and see status of BFD sessions. [#190](https://github.com/oxidecomputer/oxide.go/pull/190)

### Enhancements


### Bug fixes


### List of commits

- [428a544](https://github.com/oxidecomputer/oxide.go/commit/428a544) Update to omicron 7e0ce99 (#190)
- [a4b7143](https://github.com/oxidecomputer/oxide.go/commit/a4b7143) []NameOrID values should not be omitempty (#189)
- [b965f6a](https://github.com/oxidecomputer/oxide.go/commit/b965f6a) Tweak release checklist (#188)
- [2362321](https://github.com/oxidecomputer/oxide.go/commit/2362321) Update to Omicron 6491841 (#187)
- [8375384](https://github.com/oxidecomputer/oxide.go/commit/8375384) Rename the server field to host in the Client struct (#186)
- [6a0a73b](https://github.com/oxidecomputer/oxide.go/commit/6a0a73b) Update to Omicron 5780ff6 (#185)
- [fb28e16](https://github.com/oxidecomputer/oxide.go/commit/fb28e16) Small fix on release template (#184)
- [c9a7efb](https://github.com/oxidecomputer/oxide.go/commit/c9a7efb) Update to Omicron cc64304 (#182)
- [6a54c0b](https://github.com/oxidecomputer/oxide.go/commit/6a54c0b) Bump github.com/getkin/kin-openapi from 0.122.0 to 0.123.0 (#181)
- [172bbb1](https://github.com/oxidecomputer/oxide.go/commit/172bbb1) oxide: refactor exported client API (#180)
- [3d15f3d](https://github.com/oxidecomputer/oxide.go/commit/3d15f3d) Update to Go 1.21 (#179)
- [157d746](https://github.com/oxidecomputer/oxide.go/commit/157d746) [github] Feature request issue template (#178)
- [0dea647](https://github.com/oxidecomputer/oxide.go/commit/0dea647) [github] Add issue templates (#177)
- [642f5f4](https://github.com/oxidecomputer/oxide.go/commit/642f5f4) Update to upcoming version (#176)
- [c5e0e7e](https://github.com/oxidecomputer/oxide.go/commit/c5e0e7e) Temporarily change version to retracted (#175)
- [9e77c0e](https://github.com/oxidecomputer/oxide.go/commit/9e77c0e) Fix version retraction (#174)

# v0.1.0-beta2 (2023/Dec/18)

### Breaking changes

- **ListAll methods:** These methods now return slices instead of a pointer to a slice. [#150](https://github.com/oxidecomputer/oxide.go/pull/150)
- **Error handling:** The HTTPError type has been modified to include the HTTP response and the API's ErrorResponse type. [#145](https://github.com/oxidecomputer/oxide.go/pull/145)
- **context.Context support:** Callers are now able to specify cancellation or timeout logic. Method signatures have been modified to enable this feature. [#144](https://github.com/oxidecomputer/oxide.go/pull/144)
- **Fix generated numeric types:** Some numeric types differed to the OpenAPI spec. They are now consistent. [#142](https://github.com/oxidecomputer/oxide.go/pull/142)

### Bug fixes

- **Fix delete VPC firewall rules:** By removing `omitempty` when parsing the rules, we are able to pass an empty array to delete all firewall rules. [#158](https://github.com/oxidecomputer/terraform-provider-oxide/pull/158)

### List of commits

- [9a7cd14](https://github.com/oxidecomputer/oxide.go/commit/9a7cd14) Update version for next release (#171)
- [f95114c](https://github.com/oxidecomputer/oxide.go/commit/f95114c) Update to omicron 5827188 (#169)
- [4350767](https://github.com/oxidecomputer/oxide.go/commit/4350767) Bump github.com/getkin/kin-openapi from 0.121.0 to 0.122.0 (#163)
- [ad617b2](https://github.com/oxidecomputer/oxide.go/commit/ad617b2) Bump actions/setup-go from 4 to 5 (#160)
- [a594c9d](https://github.com/oxidecomputer/oxide.go/commit/a594c9d) Fix Makefile (#162)
- [d342cda](https://github.com/oxidecomputer/oxide.go/commit/d342cda) Update to Omicron 75cdeeb (#159)
- [fdcdc66](https://github.com/oxidecomputer/oxide.go/commit/fdcdc66) Fix VPC firewall rules delete action (#158)
- [e68d19a](https://github.com/oxidecomputer/oxide.go/commit/e68d19a) Bump github.com/getkin/kin-openapi from 0.120.0 to 0.121.0 (#154)
- [dcac177](https://github.com/oxidecomputer/oxide.go/commit/dcac177) Implement changelog automation and makefile clean up (#152)
- [ff50f82](https://github.com/oxidecomputer/oxide.go/commit/ff50f82) Retract unecessary versions (#151)
- [469b142](https://github.com/oxidecomputer/oxide.go/commit/469b142) Do not return pointer on ListAll methods (#150)
- [e20dc58](https://github.com/oxidecomputer/oxide.go/commit/e20dc58) Update SDK to Omicron f513182 (#149)
- [1c58324](https://github.com/oxidecomputer/oxide.go/commit/1c58324) Improved error handling with HTTPError type (#145)
- [9cac5e9](https://github.com/oxidecomputer/oxide.go/commit/9cac5e9) oxide: support specifying a context.Context (#144)
- [2bfa4c0](https://github.com/oxidecomputer/oxide.go/commit/2bfa4c0) Simplify detection of a list endpoint (#143)
- [772d387](https://github.com/oxidecomputer/oxide.go/commit/772d387) Fix generated numeric types (#142)
- [45e76db](https://github.com/oxidecomputer/oxide.go/commit/45e76db) Update README to reflect current methods (#141)
- [1a52f43](https://github.com/oxidecomputer/oxide.go/commit/1a52f43) Bump github.com/getkin/kin-openapi from 0.119.0 to 0.120.0 (#136)
- [7d2566a](https://github.com/oxidecomputer/oxide.go/commit/7d2566a) Bump actions/checkout from 3 to 4 (#135)

