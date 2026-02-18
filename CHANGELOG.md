# v0.8.0 (2026/Feb/17)

Generated from Oxide API version
[2026021301.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v18/rc1/openapi/nexus/nexus-2026021301.0.0-6e51ab.json)

### Breaking changes

- **Improve handling of union types.:** Use typed structs instead of `any` values to represent `oneOf` union types. [#359](https://github.com/oxidecomputer/oxide.go/pull/359) [#368](https://github.com/oxidecomputer/oxide.go/pull/368) [#376](https://github.com/oxidecomputer/oxide.go/pull/376)
- **Use functional options to construct clients.:** The `oxide.NewClient()` function now takes functional options instead of a configuration struct. [#356](https://github.com/oxidecomputer/oxide.go/pull/356)

### New features

- **Add option to disable TLS certificate verification.:** Allow the Go SDK client to skip TLS certificate verification. This is insecure and should only be used for testing or in controlled environments. [#356](https://github.com/oxidecomputer/oxide.go/pull/356)

### Enhancements

- **Expose experimental methods.:** Add methods to allow calling API endpoints that are marked as experimental. Please note that experimental endpoits are not stable and may be changed or removed without notice. [#379](https://github.com/oxidecomputer/oxide.go/pull/379)
- **Add API version header.:** Set the `API-Version` header when making API requests. [#345](https://github.com/oxidecomputer/oxide.go/pull/345)

### Bug fixes

- **:**

### List of commits

- [03aeeb2](https://github.com/oxidecomputer/oxide.go/commit/03aeeb2) changelog: add entries for v0.8.0 (#383)
- [b9d21a9](https://github.com/oxidecomputer/oxide.go/commit/b9d21a9) misc: update release checklist issue template (#384)
- [65fcc66](https://github.com/oxidecomputer/oxide.go/commit/65fcc66) Check generated code is current from ci. (#382)
- [3a77496](https://github.com/oxidecomputer/oxide.go/commit/3a77496) misc: bump omicron to `rel/v18/rc1` (#381)
- [f6fa9ab](https://github.com/oxidecomputer/oxide.go/commit/f6fa9ab) Drop redundant linters. (#378)
- [905343c](https://github.com/oxidecomputer/oxide.go/commit/905343c) Expose experimental methods. (#379)
- [30fa749](https://github.com/oxidecomputer/oxide.go/commit/30fa749) Drop flat structs for complex union types. (#376)
- [b1a40f9](https://github.com/oxidecomputer/oxide.go/commit/b1a40f9) Bump omicron version to aa12b67b. (#380)
- [10f1e97](https://github.com/oxidecomputer/oxide.go/commit/10f1e97) misc: use a separate `tools/go.mod` for `go tool` (#377)
- [44dc2f2](https://github.com/oxidecomputer/oxide.go/commit/44dc2f2) misc: bump omicron to `183a261` (#375)
- [1401cd9](https://github.com/oxidecomputer/oxide.go/commit/1401cd9) misc: bump omicron to 339fc71f (#374)
- [06dacbe](https://github.com/oxidecomputer/oxide.go/commit/06dacbe) Bump omicron to 524f11e. (#373)
- [b20dca0](https://github.com/oxidecomputer/oxide.go/commit/b20dca0) Generate proper types for untagged unions. (#368)
- [b8910e7](https://github.com/oxidecomputer/oxide.go/commit/b8910e7) chore: enable `unparam` linter and fix warnings. (#370)
- [34e197a](https://github.com/oxidecomputer/oxide.go/commit/34e197a) chore: clarify docstrings in oxide/lib.go. (#371)
- [381d09c](https://github.com/oxidecomputer/oxide.go/commit/381d09c) fix: serialize value and pointer literal enum variants the same way (#369)
- [e1ba7ef](https://github.com/oxidecomputer/oxide.go/commit/e1ba7ef) misc: bump omicron (#366)
- [a6e4781](https://github.com/oxidecomputer/oxide.go/commit/a6e4781) Use `go tool` rather than building/installing binaries. (#365)
- [ded8773](https://github.com/oxidecomputer/oxide.go/commit/ded8773) Add more helpers for oneOf types with string-like variants. (#364)
- [6cab334](https://github.com/oxidecomputer/oxide.go/commit/6cab334) Model single-discriminator, single-value enums as interfaces. (#359)
- [2f67dbc](https://github.com/oxidecomputer/oxide.go/commit/2f67dbc) Add golden tests. (#362)
- [199901c](https://github.com/oxidecomputer/oxide.go/commit/199901c) Bump golangci. (#363)
- [43715d5](https://github.com/oxidecomputer/oxide.go/commit/43715d5) Add initial agents/claude tooling. (#360)
- [e4cdc9e](https://github.com/oxidecomputer/oxide.go/commit/e4cdc9e) deps: bump omicron version (#361)
- [4c9921f](https://github.com/oxidecomputer/oxide.go/commit/4c9921f) feat: refactor client construction to use functional options (#356)
- [43d0dfc](https://github.com/oxidecomputer/oxide.go/commit/43d0dfc) feat: bump omicron version (#355)
- [71fff1f](https://github.com/oxidecomputer/oxide.go/commit/71fff1f) Refactor and test enum type creation (#347)
- [d418991](https://github.com/oxidecomputer/oxide.go/commit/d418991) Refactor: consolidate field logic in struct. (#349)
- [04fa236](https://github.com/oxidecomputer/oxide.go/commit/04fa236) fix: support versioned nexus openapi specification files (#354)
- [97a2a75](https://github.com/oxidecomputer/oxide.go/commit/97a2a75) Teach struct types to render themselves to go code. (#348)
- [24ae859](https://github.com/oxidecomputer/oxide.go/commit/24ae859) docs: add authentication section to readme. (#346)
- [00a7448](https://github.com/oxidecomputer/oxide.go/commit/00a7448) feat: support nexus api-version header (#345)
- [5196849](https://github.com/oxidecomputer/oxide.go/commit/5196849) build(deps): bump actions/checkout from 5 to 6 (#344)
- [aed681a](https://github.com/oxidecomputer/oxide.go/commit/aed681a) Handle nullable array items. (#341)
- [1866144](https://github.com/oxidecomputer/oxide.go/commit/1866144) chore: bump version (#337)

# v0.7.0 (2025/Nov/11)

Generated from Oxide API version
[20251008.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v17/rc3/openapi/nexus.json)

### Enhancements

- **Support `OXIDE_PROFILE`:** Added support for specifying the Oxide profile via the
  `OXIDE_PROFILE` environment variable. [#332](https://github.com/oxidecomputer/oxide.go/pull/332)

### Bug fixes

- **Handle `map[string]any` types:** Updated the SDK to correctly handle `map[string]any` types.
  [#325](https://github.com/oxidecomputer/oxide.go/pull/325)
- **Handle required nullable types:** Updated the SDK to correctly handle required nullable types.
  [#329](https://github.com/oxidecomputer/oxide.go/pull/329)

### List of commits

- [2a9e69d](https://github.com/oxidecomputer/oxide.go/commit/2a9e69d) chore: update changelog (#335)
- [9eaced7](https://github.com/oxidecomputer/oxide.go/commit/9eaced7) feat: bump to r17-rc3 (#334)
- [e0646c8](https://github.com/oxidecomputer/oxide.go/commit/e0646c8) oxide: bump API to 48c0c1b
  (#333)
- [a9bbb37](https://github.com/oxidecomputer/oxide.go/commit/a9bbb37) client: add support for
  OXIDE_PROFILE (#332)
- [bd541f4](https://github.com/oxidecomputer/oxide.go/commit/bd541f4) Make buildRequest method
  public. (#331)
- [b94b9b5](https://github.com/oxidecomputer/oxide.go/commit/b94b9b5) Bump actions/setup-go from 5
  to 6 (#326)
- [b8c3f7e](https://github.com/oxidecomputer/oxide.go/commit/b8c3f7e) Handle required nullable
  types. (#329)
- [02b6f67](https://github.com/oxidecomputer/oxide.go/commit/02b6f67) omicron: bump to commit
  e44ad55. (#327)
- [75e4f83](https://github.com/oxidecomputer/oxide.go/commit/75e4f83) Bump
  github.com/getkin/kin-openapi from 0.132.0 to 0.133.0 (#323)
- [3e0eed0](https://github.com/oxidecomputer/oxide.go/commit/3e0eed0) generate: handle
  map[string]any types correctly. (#325)
- [d5356df](https://github.com/oxidecomputer/oxide.go/commit/d5356df) Bump
  github.com/stretchr/testify from 1.11.0 to 1.11.1 (#324)
- [c22ab12](https://github.com/oxidecomputer/oxide.go/commit/c22ab12) Bump
  github.com/stretchr/testify from 1.10.0 to 1.11.0 (#322)
- [729a349](https://github.com/oxidecomputer/oxide.go/commit/729a349) misc: sort type properties
  before generating. (#320)
- [7037f6b](https://github.com/oxidecomputer/oxide.go/commit/7037f6b) misc: update deprecated
  actions syntax. (#321)
- [ee0ab7a](https://github.com/oxidecomputer/oxide.go/commit/ee0ab7a) oxide: bump omicron to
  71fbb2b. (#319)
- [b56a0b2](https://github.com/oxidecomputer/oxide.go/commit/b56a0b2) misc: bump to v0.7.0 (#318)

# v0.6.0 (2025/Aug/13)

Generated from Oxide API version
[20250730.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v16/rc0/openapi/nexus.json)

### Breaking changes

- **`VpcFirewallRuleProtocol` type change:** Changed the type of the `VpcFirewallRuleProtocol`
  struct to reflect upstream API changes. [#304](https://github.com/oxidecomputer/oxide.go/pull/304)

### List of commits

- [289bbe3](https://github.com/oxidecomputer/oxide.go/commit/289bbe3) Bump actions/checkout from 4
  to 5 (#312)
- [6291866](https://github.com/oxidecomputer/oxide.go/commit/6291866) oxide: update omicron version
  to dfac975. (#313)
- [ad9db82](https://github.com/oxidecomputer/oxide.go/commit/ad9db82) ci/cd: trigger on all push
  events (#315)
- [bc81674](https://github.com/oxidecomputer/oxide.go/commit/bc81674) actions: update triggers.
  (#314)
- [e5de454](https://github.com/oxidecomputer/oxide.go/commit/e5de454) oxide: optionally build from
  local openapi spec. (#311)
- [3145079](https://github.com/oxidecomputer/oxide.go/commit/3145079) Update codeowners file (#310)
- [62f5313](https://github.com/oxidecomputer/oxide.go/commit/62f5313) misc: use github team for
  codeowners (#308)
- [d8df86d](https://github.com/oxidecomputer/oxide.go/commit/d8df86d) oxide: update to omicron
  version 68a8c4b (#307)
- [34f4570](https://github.com/oxidecomputer/oxide.go/commit/34f4570) oxide: update to omicron
  version 53aea85 (#304)
- [01e729a](https://github.com/oxidecomputer/oxide.go/commit/01e729a) [bugfix] Use the correct type
  for arrays (#305)
- [83ef11f](https://github.com/oxidecomputer/oxide.go/commit/83ef11f) misc: add codeowners (#301)
- [a8e00fb](https://github.com/oxidecomputer/oxide.go/commit/a8e00fb) misc: update readme example
  code (#300)
- [a289a27](https://github.com/oxidecomputer/oxide.go/commit/a289a27) Add missing fields to network
  types (#278)
- [a65e25f](https://github.com/oxidecomputer/oxide.go/commit/a65e25f) Update to Omicron 1b4b019
  (#298)
- [c63d62d](https://github.com/oxidecomputer/oxide.go/commit/c63d62d) Bump SDK version to v0.6.0
  (#297)

# v0.5.0 (2025/Jun/11)

Generated from Oxide API version
[20250604.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v15/rc1/openapi/nexus.json)

### Breaking changes

- **Go version update:** Updated the SDK's Go version to Go 1.24. Consumers of this SDK will need to
  update to Go 1.24 as well. [#291](https://github.com/oxidecomputer/oxide.go/pull/291)
- **Set `omitzero` on specific types:** Clients can pass an empty slice and have it serialized as
  `[]`. Requires Go 1.24 or later. [#289](https://github.com/oxidecomputer/oxide.go/pull/289)

### New features

- **SiloAuthSettings:** Methods to view and update authentication settings. Namely, set token
  expiration. [#294](https://github.com/oxidecomputer/oxide.go/pull/294)
- **CurrentUserAccessToken:** Methods to view and delete a current user's auth tokens.
  [#294](https://github.com/oxidecomputer/oxide.go/pull/294)

### Bug fixes

- **Type fields:** All arrays that are nullable in the API no longer have `omitempty` to avoid
  panics if unset. [#283](https://github.com/oxidecomputer/oxide.go/pull/283)

### List of commits

- [a3144ae](https://github.com/oxidecomputer/oxide.go/commit/a3144ae) Update omicron to rc15 (#295)
- [c075870](https://github.com/oxidecomputer/oxide.go/commit/c075870) Update to omicron 760d1b0
  (#294)
- [ecfa72d](https://github.com/oxidecomputer/oxide.go/commit/ecfa72d) generate: set omitzero on
  specific types (#289)
- [855352f](https://github.com/oxidecomputer/oxide.go/commit/855352f) go: update to 1.24 (#291)
- [2e4943b](https://github.com/oxidecomputer/oxide.go/commit/2e4943b) Bump
  github.com/getkin/kin-openapi from 0.131.0 to 0.132.0 (#286)
- [65b1d0f](https://github.com/oxidecomputer/oxide.go/commit/65b1d0f) Update to omicron 5cfd735
  (#284)
- [719d3ae](https://github.com/oxidecomputer/oxide.go/commit/719d3ae) Handle nullable arrays (#283)
- [a79eb2c](https://github.com/oxidecomputer/oxide.go/commit/a79eb2c) Bump version to v0.5.0 (#281)

# v0.4.0 (2025/Apr/15)

Generated from Oxide API version
[20250409.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v14/rc1/openapi/nexus.json)

### Breaking changes

- **Integers as pointers:** All integers within the SDK's types are now `*int`. This is due to Go's
  handling of 0 as the empty value. This is specifically necessary when a field is an integer and
  also not required. [#274](https://github.com/oxidecomputer/oxide.go/pull/274)

### New features

- **Anti-affinity groups:** CRUD methods. [#269](https://github.com/oxidecomputer/oxide.go/pull/269)

### List of commits

- [0083f51](https://github.com/oxidecomputer/oxide.go/commit/0083f51) Update omicron version to
  rel/v14/rc1 (#279)
- [894605d](https://github.com/oxidecomputer/oxide.go/commit/894605d) Update to Omicron 0dad016
  (#276)
- [845061b](https://github.com/oxidecomputer/oxide.go/commit/845061b) Documentation fix (#275)
- [c8be658](https://github.com/oxidecomputer/oxide.go/commit/c8be658) Set integer fields as pointers
  (#274)
- [39db29e](https://github.com/oxidecomputer/oxide.go/commit/39db29e) Bump
  github.com/getkin/kin-openapi from 0.129.0 to 0.131.0 (#272)
- [5fd2848](https://github.com/oxidecomputer/oxide.go/commit/5fd2848) Update to omicron 8a40bb8
  (#269)
- [9d49348](https://github.com/oxidecomputer/oxide.go/commit/9d49348) Update to version v0.4.0
  (#268)

# v0.3.0 (2025/Feb/18)

Generated from Oxide API version
[20250212.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v13/rc0/openapi/nexus.json)

### New features

- **Switch Port LLDP Neighbors:** CRUD functionality for LLDP neighbors seen on a switch port.
  [#263](https://github.com/oxidecomputer/oxide.go/pull/263)

### List of commits

- [96d5f51](https://github.com/oxidecomputer/oxide.go/commit/96d5f51) Update to omicron rel/v13/rc0
  (#266)
- [6d04e20](https://github.com/oxidecomputer/oxide.go/commit/6d04e20) Update SDK to omicron e036c80
  (#263)
- [0a7b464](https://github.com/oxidecomputer/oxide.go/commit/0a7b464) Bump kin-openapi to 0.129.0
  (#262)
- [05e4807](https://github.com/oxidecomputer/oxide.go/commit/05e4807) Improve contributing docs
  (#261)
- [ea1d4d0](https://github.com/oxidecomputer/oxide.go/commit/ea1d4d0) Bump to version v0.3.0 (#259)

# v0.2.0 (2025/Jan/07)

Generated from Oxide API version
[20241204.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v12/rc1/openapi/nexus.json)

### Notes

This release is solely a version bump. Since pkg.go.dev does not recognise git tags with metadata
suffixes, we are not able to set build metadata like the previous version. See associated
[commit](https://go-review.googlesource.com/c/pkgsite/+/343631).

### List of commits

- [8357536](https://github.com/oxidecomputer/oxide.go/commit/8357536) Bump version for release
  v0.2.0 (#257)
- [f314faf](https://github.com/oxidecomputer/oxide.go/commit/f314faf) release: bump to next version
  (#255)
- [230e0cc](https://github.com/oxidecomputer/oxide.go/commit/230e0cc) release: v0.1.0+20241204.0.0
  (#254)

# v0.1.0+20241204.0.0 (2025/Jan/06)

### Breaking changes

- **Instance Update:** It's now possible to modify an instance's Memory and Ncpus count. When using
  the `InstanceUpdate` method, all parameters must be set. Otherwise, the values used will be 0.
  [247](https://github.com/oxidecomputer/oxide.go/pull/247)

### New features

- **Authenticate using Oxide credentials.toml:** Add option to authenticate using the
  `credentials.toml` file generated by the Oxide CLI.
  [244](https://github.com/oxidecomputer/oxide.go/pull/244)

### Enhancements

- **Update Go version:** The SDK's version has been updated to 1.22.
  [243](https://github.com/oxidecomputer/oxide.go/pull/243)

### List of commits

- [727dc6f](https://github.com/oxidecomputer/oxide.go/commit/727dc6f) Update to Omicron rel/v12/rc0
  (#253)
- [ed39445](https://github.com/oxidecomputer/oxide.go/commit/ed39445) Bump version to
  v0.1.0+20241204.0.0 (#252)
- [eb153ea](https://github.com/oxidecomputer/oxide.go/commit/eb153ea) Bump
  github.com/stretchr/testify from 1.9.0 to 1.10.0 (#250)
- [291b784](https://github.com/oxidecomputer/oxide.go/commit/291b784) Split long doc strings over
  multiple lines (#249)
- [6bce2f6](https://github.com/oxidecomputer/oxide.go/commit/6bce2f6) Update to Omicron 9c8aa53
  (#247)
- [18592bd](https://github.com/oxidecomputer/oxide.go/commit/18592bd) Add option to use credentials
  from CLI (#244)
- [88b2bfd](https://github.com/oxidecomputer/oxide.go/commit/88b2bfd) Makefile cleanup (#246)
- [35e937c](https://github.com/oxidecomputer/oxide.go/commit/35e937c) Update Go version to 1.22
  (#243)
- [4e5a60a](https://github.com/oxidecomputer/oxide.go/commit/4e5a60a) version bump to 0.1.0-beta10
  (#242)
- [51cac24](https://github.com/oxidecomputer/oxide.go/commit/51cac24) Bump
  github.com/getkin/kin-openapi from 0.127.0 to 0.128.0 (#239)

# v0.1.0-beta9 (2024/Oct/21)

### Breaking changes

- **OneOf generic types:** All struct field types that have different property types in the OpenAPI
  spec have now been set to `any`. [#234](https://github.com/oxidecomputer/oxide.go/pull/234)
- **NetworkingBgpAnnounceSet type:** Small change in fields.
  [236](https://github.com/oxidecomputer/oxide.go/pull/236)

### New features

- **Helper function:** New `NewPointer` function that returns a pointer to a given value.
  [235](https://github.com/oxidecomputer/oxide.go/pull/235)
- **New fields for Instance:** It is now possible to specify a boot disk and update it.
  Additionally, instances now have 'autorestart' functionality, where if set the control plane to
  automatically restart it if it enters the `Failed` state.
  [236](https://github.com/oxidecomputer/oxide.go/pull/236)
- **New types and methods:** Create, list, view and delete methods for InternetGatewayIpAddress and
  InternetGatewayIpPool. [240](https://github.com/oxidecomputer/oxide.go/pull/240)

### Bug fixes

- **Fix for fields of type `time.Time`:** Change encoding of time parameters to RFC3339.
  [232](https://github.com/oxidecomputer/oxide.go/pull/232)
- **Fix for types:** Account for additional fields 'array' types that don't specify map keys.
  [235](https://github.com/oxidecomputer/oxide.go/pull/235)

### List of commits

- [7c3ac3b](https://github.com/oxidecomputer/oxide.go/commit/7c3ac3b) Update to omicron rel/v11/rc1
  (#240)
- [92053e1](https://github.com/oxidecomputer/oxide.go/commit/92053e1) Fix nullable BootDisk field
  (#237)
- [cadd7b6](https://github.com/oxidecomputer/oxide.go/commit/cadd7b6) Update to omicron f14b561
  (#236)
- [5f5c339](https://github.com/oxidecomputer/oxide.go/commit/5f5c339) Account for additional fields
  "array" types that don't specify map keys (#235)
- [7b8deef](https://github.com/oxidecomputer/oxide.go/commit/7b8deef) Fix OneOf type templates when
  property types differ (#234)
- [2633306](https://github.com/oxidecomputer/oxide.go/commit/2633306) Change encoding of time
  parameters to RFC3339 (#232)
- [645ab82](https://github.com/oxidecomputer/oxide.go/commit/645ab82) Remove outdated checks in
  Makefile and bump tools (#230)
- [db1cf82](https://github.com/oxidecomputer/oxide.go/commit/db1cf82) Remove executable bit from
  generated source files (#229)
- [2d91c54](https://github.com/oxidecomputer/oxide.go/commit/2d91c54) Don't hardcode Bash path in
  Makefile (#228)
- [ab549ae](https://github.com/oxidecomputer/oxide.go/commit/ab549ae) Bump version to v0.1.0-beta9
  (#227)

# v0.1.0-beta8 (2024/Sep/3)

### Breaking changes

- **Enums:** All 'enum' collection variables have been changed. The word 'Collection' has been
  appended to all variable names. [#223](https://github.com/oxidecomputer/oxide.go/pull/223)
- **Instances:** The migration endpoint has been removed.
  [#223](https://github.com/oxidecomputer/oxide.go/pull/223)

### New features

- **Networking:** New BGP related methods.
  [#225](https://github.com/oxidecomputer/oxide.go/pull/225)

### Enhancements

- **Metrics:** The 'TimeseriesSchema' type now has additional fields.
  [#223](https://github.com/oxidecomputer/oxide.go/pull/223)

### List of commits

- [b4aa1b2](https://github.com/oxidecomputer/oxide.go/commit/b4aa1b2) Update to omircon
  rel/v10/rc001 (#225)
- [3ece271](https://github.com/oxidecomputer/oxide.go/commit/3ece271) Update to Omicron ede17c7 and
  refactor enum collections (#223)
- [942bccc](https://github.com/oxidecomputer/oxide.go/commit/942bccc) Bump
  github.com/getkin/kin-openapi from 0.126.0 to 0.127.0 (#222)
- [9c89a17](https://github.com/oxidecomputer/oxide.go/commit/9c89a17) Version bump (#221)

# v0.1.0-beta7 (2024/Jul/23)

### Breaking changes

- **Networking:** The `NetworkingBgpAnnounceSetCreate` method has been replaced by
  `NetworkingBgpAnnounceSetUpdate` [#218](https://github.com/oxidecomputer/oxide.go/pull/218).

### New features

- **New APIs:** Several new endpoints in [#216](https://github.com/oxidecomputer/oxide.go/pull/216)
  - VpcRouterRouteList: List routes
  - VpcRouterRouteListAllPages: List routes
  - VpcRouterRouteCreate: Create route
  - VpcRouterRouteView: Fetch route
  - VpcRouterRouteUpdate: Update route
  - VpcRouterRouteDelete: Delete route
  - VpcRouterList: List routers
  - VpcRouterListAllPages: List routers
  - VpcRouterCreate: Create VPC router
  - VpcRouterView: Fetch router
  - VpcRouterUpdate: Update router
  - VpcRouterDelete: Delete router

### List of commits

- [3682a00](https://github.com/oxidecomputer/oxide.go/commit/3682a00) Update to omicron bedb238
  (#218)
- [c52f6e0](https://github.com/oxidecomputer/oxide.go/commit/c52f6e0) Bump
  github.com/getkin/kin-openapi from 0.125.0 to 0.126.0 (#217)
- [06dd780](https://github.com/oxidecomputer/oxide.go/commit/06dd780) Update to Omicron 97fe552
  (#216)
- [e44fdd5](https://github.com/oxidecomputer/oxide.go/commit/e44fdd5) Bump
  github.com/getkin/kin-openapi from 0.124.0 to 0.125.0 (#215)
- [4151b01](https://github.com/oxidecomputer/oxide.go/commit/4151b01) Version bump (#214)

# v0.1.0-beta6 (2024/May/9)

### Breaking changes

- **Types:** Changes to BGP related types.
  [#212](https://github.com/oxidecomputer/oxide.go/pull/212)

### List of commits

- [a4018ce](https://github.com/oxidecomputer/oxide.go/commit/a4018ce) Update to omicron c1f9e8f
  (#212)
- [bb16ad2](https://github.com/oxidecomputer/oxide.go/commit/bb16ad2) Version bump (#210)

# v0.1.0-beta5 (2024/May/6)

### New features

- **New APIs:** Several new endpoints in [#208](https://github.com/oxidecomputer/oxide.go/pull/208)
  - NetworkingAllowListView: Get user-facing services IP allowlist
  - NetworkingAllowListUpdate: Update user-facing services IP allowlist
  - NetworkingSwitchPortStatus: Get switch port status

### List of commits

- [75ad608](https://github.com/oxidecomputer/oxide.go/commit/75ad608) Update to omicron f2602b5
  (#208)
- [44a6751](https://github.com/oxidecomputer/oxide.go/commit/44a6751) Update kin-openapi to 0.124.0
  (#206)
- [b4e284c](https://github.com/oxidecomputer/oxide.go/commit/b4e284c) Version bump (#204)

# v0.1.0-beta4 (2024/Apr/3)

### New features

- **New API endpoints:** Floating IP update, IP pool utilization view, physical disk view,
  timeseries query, timeseries schema list, and BGP message history.
  [#195](https://github.com/oxidecomputer/oxide.go/pull/195),
  [#201](https://github.com/oxidecomputer/oxide.go/pull/201),
  [#202](https://github.com/oxidecomputer/oxide.go/pull/202)

### Enhancements

- **Documentation:** Go doc comments now include which fields are required for each type.
  [#198](https://github.com/oxidecomputer/oxide.go/pull/198)

### List of commits

- [f488d8e](https://github.com/oxidecomputer/oxide.go/commit/f488d8e) Update to omicron afb2e9a
  (#202)
- [f7d1056](https://github.com/oxidecomputer/oxide.go/commit/f7d1056) Update to omicron a3fa540
  (#201)
- [35ead62](https://github.com/oxidecomputer/oxide.go/commit/35ead62) Bump
  softprops/action-gh-release from 1 to 2 (#199)
- [8359042](https://github.com/oxidecomputer/oxide.go/commit/8359042) Document required fields
  (#198)
- [2d221d4](https://github.com/oxidecomputer/oxide.go/commit/2d221d4) Remove unecessary env var from
  GH action (#197)
- [9b0cf8d](https://github.com/oxidecomputer/oxide.go/commit/9b0cf8d) Bump
  github.com/stretchr/testify from 1.8.4 to 1.9.0 (#196)
- [043c873](https://github.com/oxidecomputer/oxide.go/commit/043c873) Update SDK to Omicron dcd3d9e
  (#195)
- [20c490d](https://github.com/oxidecomputer/oxide.go/commit/20c490d) Write correct date on
  changelog (#193)
- [38e6c01](https://github.com/oxidecomputer/oxide.go/commit/38e6c01) Bump version for next release
  (#192)

# v0.1.0-beta3 (2024/Feb/13)

### Breaking changes

- **Go version:** Minimum required Go version has been updated to 1.21.
  [#179](https://github.com/oxidecomputer/oxide.go/pull/179)
- **NewClient API change:** The `NewClient` function has been updated to no longer require a user
  agent parameter. [#180](https://github.com/oxidecomputer/oxide.go/pull/180)
- **NewClientFromEnv removal:** The `NewClientFromEnv` function has been removed. Users should use
  `NewClient` instead. [#180](https://github.com/oxidecomputer/oxide.go/pull/180)
- **Method renames:** Several methods have had slight name changes to better reflect their
  functionality. [#182](https://github.com/oxidecomputer/oxide.go/pull/182)
- **Types:** Several types have added fields and/or renames.
  [#182](https://github.com/oxidecomputer/oxide.go/pull/182),
  [#185](https://github.com/oxidecomputer/oxide.go/pull/185),
  [#190](https://github.com/oxidecomputer/oxide.go/pull/190)

### New features

- **New instance APIs:** Live attach and detach of external IPs to an instance.
  [#182](https://github.com/oxidecomputer/oxide.go/pull/182)
- **New IP pool APIs:** Several silo IP pool maintenance endpoints.
  [#182](https://github.com/oxidecomputer/oxide.go/pull/182),
  [#187](https://github.com/oxidecomputer/oxide.go/pull/187)
- **New SSH keys APIs:** Endpoint to list SSH keys that were added to an instance on create.
  [#185](https://github.com/oxidecomputer/oxide.go/pull/185)
- **New networking APIs:** Enable, disable and see status of BFD sessions.
  [#190](https://github.com/oxidecomputer/oxide.go/pull/190)

### List of commits

- [428a544](https://github.com/oxidecomputer/oxide.go/commit/428a544) Update to omicron 7e0ce99
  (#190)
- [a4b7143](https://github.com/oxidecomputer/oxide.go/commit/a4b7143) []NameOrID values should not
  be omitempty (#189)
- [b965f6a](https://github.com/oxidecomputer/oxide.go/commit/b965f6a) Tweak release checklist (#188)
- [2362321](https://github.com/oxidecomputer/oxide.go/commit/2362321) Update to Omicron 6491841
  (#187)
- [8375384](https://github.com/oxidecomputer/oxide.go/commit/8375384) Rename the server field to
  host in the Client struct (#186)
- [6a0a73b](https://github.com/oxidecomputer/oxide.go/commit/6a0a73b) Update to Omicron 5780ff6
  (#185)
- [fb28e16](https://github.com/oxidecomputer/oxide.go/commit/fb28e16) Small fix on release template
  (#184)
- [c9a7efb](https://github.com/oxidecomputer/oxide.go/commit/c9a7efb) Update to Omicron cc64304
  (#182)
- [6a54c0b](https://github.com/oxidecomputer/oxide.go/commit/6a54c0b) Bump
  github.com/getkin/kin-openapi from 0.122.0 to 0.123.0 (#181)
- [172bbb1](https://github.com/oxidecomputer/oxide.go/commit/172bbb1) oxide: refactor exported
  client API (#180)
- [3d15f3d](https://github.com/oxidecomputer/oxide.go/commit/3d15f3d) Update to Go 1.21 (#179)
- [157d746](https://github.com/oxidecomputer/oxide.go/commit/157d746) [github] Feature request issue
  template (#178)
- [0dea647](https://github.com/oxidecomputer/oxide.go/commit/0dea647) [github] Add issue templates
  (#177)
- [642f5f4](https://github.com/oxidecomputer/oxide.go/commit/642f5f4) Update to upcoming version
  (#176)
- [c5e0e7e](https://github.com/oxidecomputer/oxide.go/commit/c5e0e7e) Temporarily change version to
  retracted (#175)
- [9e77c0e](https://github.com/oxidecomputer/oxide.go/commit/9e77c0e) Fix version retraction (#174)

# v0.1.0-beta2 (2023/Dec/18)

### Breaking changes

- **ListAll methods:** These methods now return slices instead of a pointer to a slice.
  [#150](https://github.com/oxidecomputer/oxide.go/pull/150)
- **Error handling:** The HTTPError type has been modified to include the HTTP response and the
  API's ErrorResponse type. [#145](https://github.com/oxidecomputer/oxide.go/pull/145)
- **context.Context support:** Callers are now able to specify cancellation or timeout logic. Method
  signatures have been modified to enable this feature.
  [#144](https://github.com/oxidecomputer/oxide.go/pull/144)
- **Fix generated numeric types:** Some numeric types differed to the OpenAPI spec. They are now
  consistent. [#142](https://github.com/oxidecomputer/oxide.go/pull/142)

### Bug fixes

- **Fix delete VPC firewall rules:** By removing `omitempty` when parsing the rules, we are able to
  pass an empty array to delete all firewall rules.
  [#158](https://github.com/oxidecomputer/oxide.go/pull/158)

### List of commits

- [9a7cd14](https://github.com/oxidecomputer/oxide.go/commit/9a7cd14) Update version for next
  release (#171)
- [f95114c](https://github.com/oxidecomputer/oxide.go/commit/f95114c) Update to omicron 5827188
  (#169)
- [4350767](https://github.com/oxidecomputer/oxide.go/commit/4350767) Bump
  github.com/getkin/kin-openapi from 0.121.0 to 0.122.0 (#163)
- [ad617b2](https://github.com/oxidecomputer/oxide.go/commit/ad617b2) Bump actions/setup-go from 4
  to 5 (#160)
- [a594c9d](https://github.com/oxidecomputer/oxide.go/commit/a594c9d) Fix Makefile (#162)
- [d342cda](https://github.com/oxidecomputer/oxide.go/commit/d342cda) Update to Omicron 75cdeeb
  (#159)
- [fdcdc66](https://github.com/oxidecomputer/oxide.go/commit/fdcdc66) Fix VPC firewall rules delete
  action (#158)
- [e68d19a](https://github.com/oxidecomputer/oxide.go/commit/e68d19a) Bump
  github.com/getkin/kin-openapi from 0.120.0 to 0.121.0 (#154)
- [dcac177](https://github.com/oxidecomputer/oxide.go/commit/dcac177) Implement changelog automation
  and makefile clean up (#152)
- [ff50f82](https://github.com/oxidecomputer/oxide.go/commit/ff50f82) Retract unecessary versions
  (#151)
- [469b142](https://github.com/oxidecomputer/oxide.go/commit/469b142) Do not return pointer on
  ListAll methods (#150)
- [e20dc58](https://github.com/oxidecomputer/oxide.go/commit/e20dc58) Update SDK to Omicron f513182
  (#149)
- [1c58324](https://github.com/oxidecomputer/oxide.go/commit/1c58324) Improved error handling with
  HTTPError type (#145)
- [9cac5e9](https://github.com/oxidecomputer/oxide.go/commit/9cac5e9) oxide: support specifying a
  context.Context (#144)
- [2bfa4c0](https://github.com/oxidecomputer/oxide.go/commit/2bfa4c0) Simplify detection of a list
  endpoint (#143)
- [772d387](https://github.com/oxidecomputer/oxide.go/commit/772d387) Fix generated numeric types
  (#142)
- [45e76db](https://github.com/oxidecomputer/oxide.go/commit/45e76db) Update README to reflect
  current methods (#141)
- [1a52f43](https://github.com/oxidecomputer/oxide.go/commit/1a52f43) Bump
  github.com/getkin/kin-openapi from 0.119.0 to 0.120.0 (#136)
- [7d2566a](https://github.com/oxidecomputer/oxide.go/commit/7d2566a) Bump actions/checkout from 3
  to 4 (#135)
