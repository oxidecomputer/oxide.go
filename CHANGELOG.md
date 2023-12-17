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

