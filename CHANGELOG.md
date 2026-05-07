# Changelog

## [0.10.0](https://github.com/KeeperHub/cli/compare/v0.9.0...v0.10.0) (2026-05-07)


### Features

* add `kh w feedback` thin wrapper ([79be12c](https://github.com/KeeperHub/cli/commit/79be12c3ffcedee7930007728a302b59cb349aec))
* KEEP-438 add Cloudflare Access header support ([c5f1326](https://github.com/KeeperHub/cli/commit/c5f132627453ea6d1fb2aa5460bf445657a07443))
* KEEP-438 add Cloudflare Access header support for gated environments ([35c2802](https://github.com/KeeperHub/cli/commit/35c2802c4b1f7a02c92eb668b164c4567133214c))


### Bug Fixes

* KEEP-438 append CF_Authorization to existing Cookie header instead of clobbering ([e4b4f92](https://github.com/KeeperHub/cli/commit/e4b4f927e64be58f24484d2a06356ceb38b919ef))

## [0.9.0](https://github.com/KeeperHub/cli/compare/v0.8.0...v0.9.0) (2026-04-21)


### Features

* **35:** add kh wallet add/info/fund/link wrappers around @keeperhub/wallet ([aaaadff](https://github.com/KeeperHub/cli/commit/aaaadff7981705b320b6fa299f4f0a9b5a91ea5c))
* kh wallet add/info/fund/link wrappers around @keeperhub/wallet ([100febd](https://github.com/KeeperHub/cli/commit/100febdda1a757fcb7c6667ec82df506bce68885))

## [0.8.0](https://github.com/KeeperHub/cli/compare/v0.7.0...v0.8.0) (2026-03-25)


### Features

* **KEEP-143:** add kh read, chain list, and remove auth wall for public commands ([d4abdf9](https://github.com/KeeperHub/cli/commit/d4abdf9d65dd3617dd74f3c288edf8de7ccf73c2))
* **KEEP-143:** add kh read, kh plugin, kh chain, and remove auth wall ([59ca228](https://github.com/KeeperHub/cli/commit/59ca22808c765c0c43aeaa7b8e19d30b28d69bb8))


### Bug Fixes

* match chain API field names (defaultPrimaryRpc, chainType, isEnabled) ([2f54353](https://github.com/KeeperHub/cli/commit/2f543533b6b46517a24cf16eaa0d5fda05682062))
* parse protocol list from actions map instead of plugins array ([378c41e](https://github.com/KeeperHub/cli/commit/378c41e62cd87c6e78a6ce43c8addbf17162f79e))

## [0.7.0](https://github.com/KeeperHub/cli/compare/v0.6.0...v0.7.0) (2026-03-25)


### Features

* **KEEP-145:** add --org flag for organization override ([bbd5b76](https://github.com/KeeperHub/cli/commit/bbd5b767f600ebafa841a9a3a23adfe8ba2451a9))
* **KEEP-145:** add --org flag for organization override ([ea1eb98](https://github.com/KeeperHub/cli/commit/ea1eb9839d4f8226588d17e1c22a9cf716f5b827))


### Bug Fixes

* strip organizationId from POST body and resolve header ([c6da13b](https://github.com/KeeperHub/cli/commit/c6da13b54571898e2ef894798ee4024347893ec9))
* update docs sync target from techops-services to KeeperHub org ([6e8d92f](https://github.com/KeeperHub/cli/commit/6e8d92f161747fbbe9971cfc198ae91ac6c18d5d))
* update docs sync target from techops-services to KeeperHub org ([07741a1](https://github.com/KeeperHub/cli/commit/07741a166b53d91223c9f4e6cefc776734c40a71))

## [0.6.0](https://github.com/KeeperHub/cli/compare/v0.5.3...v0.6.0) (2026-03-23)


### Features

* add --force flag to workflow delete command ([746840a](https://github.com/KeeperHub/cli/commit/746840a902b75232004949038d539ed76e6dda39))
* add --force flag to workflow delete command ([7da2051](https://github.com/KeeperHub/cli/commit/7da20510d74fc1c104502fbacb7901f898a4655c))
* add .deb and .rpm package support for Linux ([7714e74](https://github.com/KeeperHub/cli/commit/7714e74ee346068c938fc0d7b9ab9face7218487))

## [0.5.3](https://github.com/KeeperHub/cli/compare/v0.5.2...v0.5.3) (2026-03-16)


### Bug Fixes

* resolve --host flag in MCP serve mode ([aabc8db](https://github.com/KeeperHub/cli/commit/aabc8db1ebd9ae3f17cd31842127c8bef52b2583))
* resolve --host flag in MCP serve mode ([7435f64](https://github.com/KeeperHub/cli/commit/7435f6425b577195dc4c2fb78120ebd3bb370e5a))
* set git identity in sync-cli-docs workflow ([479c3cc](https://github.com/KeeperHub/cli/commit/479c3ccd4366c84b5be993b589e2bde6735b1aa5))
* set git identity in sync-cli-docs workflow ([748ecd2](https://github.com/KeeperHub/cli/commit/748ecd2ee52a7335f6b669dddf6f1d1f9cb2342b))

## [0.5.2](https://github.com/KeeperHub/cli/compare/v0.5.1...v0.5.2) (2026-03-16)


### Bug Fixes

* multi-host auth with scheme-aware host resolution ([78d76f0](https://github.com/KeeperHub/cli/commit/78d76f0937438ed78a023d5b1875ff87b20d6d67))
* multi-host auth with scheme-aware host resolution ([bf64626](https://github.com/KeeperHub/cli/commit/bf64626f722817745de2e6bffa065f7ba49b9aed))

## [0.5.1](https://github.com/KeeperHub/cli/compare/v0.5.0...v0.5.1) (2026-03-15)


### Bug Fixes

* docs-sync workflow_dispatch trigger ([9dc2605](https://github.com/KeeperHub/cli/commit/9dc2605e63e7cb2b50a4f3f227720d6e2ad9a30b))
* fix YAML syntax error in docs sync workflow ([fb83447](https://github.com/KeeperHub/cli/commit/fb834472ee1c16a3b1e30e99fe4d07c02eac2948))
* remove blank line in workflow YAML that breaks parser ([783e989](https://github.com/KeeperHub/cli/commit/783e98940919b48b40f1f984de2c588e322a1a48))
* remove empty braces from workflow_dispatch trigger ([eeb7957](https://github.com/KeeperHub/cli/commit/eeb7957cfc5639776303976995d963f85d35389f))
* rename docs-sync workflow to fix workflow_dispatch ([be3955c](https://github.com/KeeperHub/cli/commit/be3955c79368e6ef7f1016573e6daf1d8908de92))
* rename docs-sync workflow to force new GitHub workflow ID ([9bcbbfc](https://github.com/KeeperHub/cli/commit/9bcbbfc64d1cd913612c441106a6ed1020ea85a4))
* reorder docs-sync triggers to fix workflow_dispatch ([4dc3ed2](https://github.com/KeeperHub/cli/commit/4dc3ed249d7105f2019b8bdf4e1f5f234ec45b76))
* reorder docs-sync triggers to fix workflow_dispatch ([22d4a6b](https://github.com/KeeperHub/cli/commit/22d4a6b2d2c40975e0522144d1d47594dfe28bc6))

## [0.5.0](https://github.com/KeeperHub/cli/compare/v0.4.1...v0.5.0) (2026-03-15)


### Features

* add workflow create, delete, update commands and MCP workflow tools ([e36110d](https://github.com/KeeperHub/cli/commit/e36110d2a22d99aaf8d2e6271ded81fbf8c5ca73))
* add workflow CRUD commands and MCP workflow tools ([0a174d4](https://github.com/KeeperHub/cli/commit/0a174d4addbcd5d97cd057203262df10eb409e9f))


### Bug Fixes

* ensure create request always includes nodes and edges arrays ([ef27c06](https://github.com/KeeperHub/cli/commit/ef27c06263cb250cb488cc1061bbe1f627856538))
* show clear error when deleting workflow with existing runs ([5fe4422](https://github.com/KeeperHub/cli/commit/5fe442270a41e0ba135bca4a7a59e278b60bd735))

## [0.4.1](https://github.com/KeeperHub/cli/compare/v0.4.0...v0.4.1) (2026-03-14)


### Bug Fixes

* embed version from release-please manifest ([43743d4](https://github.com/KeeperHub/cli/commit/43743d4e69c916afafef99afc10a6026bbd10e8a))
* embed version from release-please manifest so kh update doesn't panic ([958ea26](https://github.com/KeeperHub/cli/commit/958ea2658fb4f5aa9dd31ed0114a6fce2ad68509))

## [0.4.0](https://github.com/KeeperHub/cli/compare/v0.3.0...v0.4.0) (2026-03-13)


### Features

* CLI output overhaul and bug fixes ([6c9fc3f](https://github.com/KeeperHub/cli/commit/6c9fc3f2a35e0977ef1258f155da7ba8e11af6ef))
* CLI output overhaul and bug fixes ([9961bee](https://github.com/KeeperHub/cli/commit/9961bee82b919833d715f9330df35cd09de01ce6))

## [0.3.0](https://github.com/KeeperHub/cli/compare/v0.2.3...v0.3.0) (2026-03-13)


### Features

* browser-based OAuth via server-side relay with nonce ([8bbb5fb](https://github.com/KeeperHub/cli/commit/8bbb5fb4a304f9f0ebfd4091b0683b3701306188))
* browser-based OAuth via server-side relay with nonce ([5635884](https://github.com/KeeperHub/cli/commit/5635884c62350c0cae73724767b099aebb549ad5))
* make device flow the default auth method ([eb0b8ff](https://github.com/KeeperHub/cli/commit/eb0b8ffa40a79c45042f468f041c9a163f9be0b8))


### Bug Fixes

* add Origin header to device flow requests for Better Auth CSRF ([9ff8747](https://github.com/KeeperHub/cli/commit/9ff8747eea2992dee541ddd618989cb9c9f4a429))

## [0.2.3](https://github.com/KeeperHub/cli/compare/v0.2.2...v0.2.3) (2026-03-13)


### Bug Fixes

* change default host from app.keeperhub.io to app.keeperhub.com ([b90b587](https://github.com/KeeperHub/cli/commit/b90b587fbb5f0c8b9f8ccaca527afe1945cdf695))
* change default host from app.keeperhub.io to app.keeperhub.com ([ed08236](https://github.com/KeeperHub/cli/commit/ed08236e3d0c278d5ab7ec9ae7735bd58ce608d7))
* set homebrew cask name to kh ([574e7ba](https://github.com/KeeperHub/cli/commit/574e7baf1c8663396e9f67b59041c4394b90132b))
* set homebrew cask name to kh instead of defaulting to cli ([9c86036](https://github.com/KeeperHub/cli/commit/9c86036a133665e1fe394a1bb5e006a54d4657b7))

## [0.2.2](https://github.com/KeeperHub/cli/compare/v0.2.1...v0.2.2) (2026-03-13)


### Bug Fixes

* remove ([b12c5db](https://github.com/KeeperHub/cli/commit/b12c5db5f9f8a92f8dbcaeb4d1142d65e85663b5))

## [0.2.1](https://github.com/KeeperHub/cli/compare/v0.2.0...v0.2.1) (2026-03-13)


### Bug Fixes

* **ci:** exclude fmt.Fprint* and Body.Close from errcheck ([b4f8bd3](https://github.com/KeeperHub/cli/commit/b4f8bd31168b2f08d86ea19ccc89ee2687b52bba))
* **ci:** fix remaining ST1005 errors and update test assertion ([986f7d0](https://github.com/KeeperHub/cli/commit/986f7d0d0b124822f135836f3984be440daf9a45))
* **ci:** remove defunct gosimple linter and fix docs-check working dir ([783ae20](https://github.com/KeeperHub/cli/commit/783ae204b435fd4122fb656912b0ef042971f043))
* **ci:** resolve all lint errors across codebase ([ee08b00](https://github.com/KeeperHub/cli/commit/ee08b009d3338bcdd3cb475493f439af0296e0f0))
* **ci:** update golangci-lint config and regenerate docs ([5fbaece](https://github.com/KeeperHub/cli/commit/5fbaecee910cae944b7aa4bc4d46eac1cc174ea8))
* **ci:** update golangci-lint config for v2 schema and regenerate docs ([c02fbe5](https://github.com/KeeperHub/cli/commit/c02fbe504db5fe7409747c9206b5cf2aada0cf36))
* **ci:** use default:none and enable only core linters ([44b2555](https://github.com/KeeperHub/cli/commit/44b25557f46b3781e2e9f37b280f6107bf6878cf))
* **ci:** use PAT in release-please to trigger goreleaser ([b8f88d5](https://github.com/KeeperHub/cli/commit/b8f88d5b7878718a94809c250e2ddf6550a70b3f))
* **ci:** use PAT in release-please to trigger goreleaser on tags ([3c7b645](https://github.com/KeeperHub/cli/commit/3c7b64576089b01046ac9529c7fc1557a73cc0f8))

## [0.2.0](https://github.com/KeeperHub/cli/compare/v0.1.0...v0.2.0) (2026-03-13)


### Features

* **19-01:** add config system with XDG paths and multi-host hosts.yml ([55e5b3a](https://github.com/KeeperHub/cli/commit/55e5b3a589535d8f931089dcc6b42ed89f56547e))
* **19-01:** initialize Go module, IOStreams, Factory, version, and error types ([bb3d3c1](https://github.com/KeeperHub/cli/commit/bb3d3c1e317bef8490f5110bba057757e1bbb328))
* **19-02:** retryable HTTP client with version header middleware ([e0f2210](https://github.com/KeeperHub/cli/commit/e0f22107eb5069e342604ca34db2ea0156a25b32))
* **19-02:** root command with global flags and updated main.go ([ce6fa75](https://github.com/KeeperHub/cli/commit/ce6fa753b1eb34437af07503c22fe87e5a38ddbf))
* **19-03:** auth, config, serve, version, doctor, and completion commands ([0baf956](https://github.com/KeeperHub/cli/commit/0baf956a83041221b156e72ddd6d2d3f30c9a4b1))
* **19-03:** workflow, run, and execute command groups with stubs ([51f154a](https://github.com/KeeperHub/cli/commit/51f154afe5900f4f0c5fda52795dbd5012d816aa))
* **19-04:** add action, protocol, wallet, template, and billing command stubs ([68be55c](https://github.com/KeeperHub/cli/commit/68be55cd625f54d80ffcec4eb8055a6b8014662b))
* **19-04:** add project, tag, api-key, and org command stubs ([fd1a84a](https://github.com/KeeperHub/cli/commit/fd1a84a5f2bb7ef35556895ed30655ff0807a49b))
* **19-05:** wire all commands and add integration tests ([fa54bbb](https://github.com/KeeperHub/cli/commit/fa54bbbdbe78baea87fecf5fc660e9437fbc94cd))
* **20-02:** add error types, exit code handler, and 429 retry guard ([6b96cf6](https://github.com/KeeperHub/cli/commit/6b96cf6bd3a57906726a37ddfff7ef1b32bda279))
* **20-02:** add JSON, jq, and table output modules with tests ([a0b1ca6](https://github.com/KeeperHub/cli/commit/a0b1ca622e1758869297a5cf82895f2065035244))
* **20-02:** add Printer that unifies --json/--jq/table output routing ([b1af1d0](https://github.com/KeeperHub/cli/commit/b1af1d02b7c24ca0d35a5f1d384587585efe31a4))
* **20-03:** browser OAuth flow and device code flow ([be70f90](https://github.com/KeeperHub/cli/commit/be70f9005cd6d8cc0e82b7d9bc2c59a00dd59ba5))
* **20-03:** keyring token storage and resolution chain ([509dab5](https://github.com/KeeperHub/cli/commit/509dab55975f4fc4b1c0b196fc455a46fd854ff3))
* **20-04:** implement auth login, logout, and status commands ([f01a487](https://github.com/KeeperHub/cli/commit/f01a4877195e1d03ebfc5fb43d9cc39e411208d8))
* **20-04:** wire auth.ResolveToken into HTTPClient factory ([122aaf2](https://github.com/KeeperHub/cli/commit/122aaf243c8510918e0511be3ee18094a3100c06))
* **21-01:** implement workflow list and get commands with tests ([ba0dbd4](https://github.com/KeeperHub/cli/commit/ba0dbd4890d2055a9135a623df1b03adb24beecc))
* **21-01:** remove apikey and check-and-execute command stubs ([9acd41d](https://github.com/KeeperHub/cli/commit/9acd41df4af7684720fa207c59ecbe5c6f79691b))
* **21-02:** implement workflow go-live and pause commands with tests ([2c93b6d](https://github.com/KeeperHub/cli/commit/2c93b6d93261bae2a1be6077b1a9045561776fee))
* **21-02:** implement workflow run command with fire-and-forget and --wait polling ([d93ea48](https://github.com/KeeperHub/cli/commit/d93ea48c689918d426bc141083b60d85cdadf8d2))
* **21-03:** implement run logs and cancel commands ([8555912](https://github.com/KeeperHub/cli/commit/855591278c8fb248d3748257c1752fbac7b20967))
* **21-03:** implement run status command with --watch live polling ([2aa876f](https://github.com/KeeperHub/cli/commit/2aa876f87fe2c4716efcef48636a62ed3f39274c))
* **21-04:** implement execution status command ([9625852](https://github.com/KeeperHub/cli/commit/9625852e1ede8a28fc7a4c64333ad8e6b77dc01a))
* **21-04:** implement transfer and contract-call commands ([8baae2d](https://github.com/KeeperHub/cli/commit/8baae2d810014bb80459fd7d44273f86c40b0ffa))
* **21-05:** apply BuildBaseURL to all workflow, execute, and auth files ([1d19b1e](https://github.com/KeeperHub/cli/commit/1d19b1e75e22aaa2fef10bf031114cb0a012516a))
* **21-05:** extract BuildBaseURL to internal/http/url.go with tests ([07f328a](https://github.com/KeeperHub/cli/commit/07f328a35fe460c81ab8a433163d9640399c9980))
* **21-06:** add config.yml fallback to ActiveHost ([7f96f96](https://github.com/KeeperHub/cli/commit/7f96f96bf2541cb099e88d753bd849ca523d9096))
* **22-01:** implement project CRUD commands ([62e0bff](https://github.com/KeeperHub/cli/commit/62e0bff41784332aee70f83fe978a0089c74db3e))
* **22-01:** implement tag CRUD commands ([62419e8](https://github.com/KeeperHub/cli/commit/62419e8cce4eb4a4a26b6a7415c68824e48b5902))
* **22-02:** implement org list and switch commands ([feaa214](https://github.com/KeeperHub/cli/commit/feaa2148c30cfbbd5e64a76564c6c0e87162bcaa))
* **22-02:** implement org members command ([ead80ff](https://github.com/KeeperHub/cli/commit/ead80ffb2e0c33e99807da7e3b8887541cfa814a))
* **22-03:** implement cache infrastructure package ([6ded9a3](https://github.com/KeeperHub/cli/commit/6ded9a35caa516f1ea77f6108cc72ae475d921b4))
* **22-03:** implement protocol list and get commands with cache ([5f7a5cf](https://github.com/KeeperHub/cli/commit/5f7a5cff62c2e6d8def626b84e6d0d4323aa9c83))
* **22-04:** implement action list and get commands ([d54dac7](https://github.com/KeeperHub/cli/commit/d54dac74aab56371832ea0e30ac0c9f58f2f6f7d))
* **22-04:** implement wallet balance and tokens commands ([34534d1](https://github.com/KeeperHub/cli/commit/34534d1964a44331163120df27e7f7dad546a551))
* **22-05:** implement billing status and usage commands ([85947a6](https://github.com/KeeperHub/cli/commit/85947a62d6e21b80502e6683453d17f8b516d1ee))
* **22-05:** implement template list and deploy commands ([497667a](https://github.com/KeeperHub/cli/commit/497667a8ef7149c88b6cd541a1c8f031a3ba599c))
* **22-06:** implement doctor command with 6 parallel health checks ([fb1b84c](https://github.com/KeeperHub/cli/commit/fb1b84ccd97daff533a6e3868b8d5e88dade8bed))
* **23-01:** export RegisterTools, BuildInputSchema, MakeToolHandler for testability ([cf4db79](https://github.com/KeeperHub/cli/commit/cf4db7987d84419d3d13f1f9556cbbf7bf66ba63))
* **23-01:** implement MCP server with stdout isolation, schema types, and dynamic tool registration ([ae05a76](https://github.com/KeeperHub/cli/commit/ae05a764c680f41d5e334d51bd43c9d73fa4dc3c))
* **23-02:** add Example, Long, and See also help text to all commands ([f081003](https://github.com/KeeperHub/cli/commit/f081003f2aabc8be112dc829d6f090a0c0e5df3d))
* **23-02:** create help topic commands for environment, exit-codes, and formatting ([ddba326](https://github.com/KeeperHub/cli/commit/ddba3267e7f822762897d0b488f2cf8ad2a88467))
* **23-03:** add integration test skeletons and CI docs-check/integration jobs ([74cb7e8](https://github.com/KeeperHub/cli/commit/74cb7e818634aadce870a6023aec5bfc069a3685))
* **24-01:** implement kh update command with Homebrew detection and self-update ([07a012d](https://github.com/KeeperHub/cli/commit/07a012d153e2476889e5dd90f17a2e5c998fde0e))
* **24-02:** add completions script and update GoReleaser config with homebrew_casks ([e2144db](https://github.com/KeeperHub/cli/commit/e2144db4f261fc240c2363ae4e36491c88277123))
