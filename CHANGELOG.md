# Changelog

## [0.1.0](https://github.com/jamesjohnsdev/issues/compare/v0.0.4...v0.1.0) (2026-06-21)


### Features

* able to close with comment ([bf7562c](https://github.com/jamesjohnsdev/issues/commit/bf7562c7554111df20e5de2bec054cf5f368a3ab))
* added `--web` flag to open browser view ([e075e9f](https://github.com/jamesjohnsdev/issues/commit/e075e9fba138355af18ce1a0f3e75a315e599e69))
* cli commands for issue comments ([df364b2](https://github.com/jamesjohnsdev/issues/commit/df364b2f2497d3c0008feafe9abbd97cc47b195a))
* comment compatability ([14c5d6f](https://github.com/jamesjohnsdev/issues/commit/14c5d6f8bcc60704115fcd5979ddd6aae5b92384))
* enable metadata support for comments ([a8f1fde](https://github.com/jamesjohnsdev/issues/commit/a8f1fde6eb1f46686859309d0ad25811ed2e2bc1))
* merge issues together and mark duplicates ([40f2ca8](https://github.com/jamesjohnsdev/issues/commit/40f2ca8248b9893ef1b4f438e743fd67e61842b5))
* run issues commands from subdirectories ([5c74a47](https://github.com/jamesjohnsdev/issues/commit/5c74a4714b96268b31b435a8f5810125dddbd3f4))


### Bug Fixes

* guard against empty/whitespace editor env and exclusive flags ([439acf1](https://github.com/jamesjohnsdev/issues/commit/439acf1fcf97d63ec0d8fe1d611401128bbd57ee))


### Performance Improvements

* **sync:** skip unchanged issues and parallelize comment fetches ([b8076c7](https://github.com/jamesjohnsdev/issues/commit/b8076c7b2c4ce66dda064fddc81c1b2ad7dc93b6))

## [0.0.4](https://github.com/jamesjohnsdev/issues/compare/v0.0.3...v0.0.4) (2026-06-18)


### Bug Fixes

* editor opens when passing through e ([b85c97f](https://github.com/jamesjohnsdev/issues/commit/b85c97f633fd33a665edecd5ac96963ddca6ecee))

## [0.0.3](https://github.com/jamesjohnsdev/issues/compare/v0.0.2...v0.0.3) (2026-06-18)


### Features

* **create:** show full schema and body marker in editor ([efce93b](https://github.com/jamesjohnsdev/issues/commit/efce93b0eb95eb27813231fece7f21a09bcc7971)), closes [#7](https://github.com/jamesjohnsdev/issues/issues/7)


### Bug Fixes

* unchecked errors ([83af814](https://github.com/jamesjohnsdev/issues/commit/83af814c17a0bc68c45055b5182930997070b803))

## [0.0.2](https://github.com/jamesjohnsdev/issues/compare/v0.0.1...v0.0.2) (2026-06-18)


### Features

* **create:** add stepwise interactive prompts when no args given ([1e54bc5](https://github.com/jamesjohnsdev/issues/commit/1e54bc551e5ea20321b820bad8609955728c9f49))

## [0.0.1](https://github.com/jamesjohnsdev/issues/compare/v0.0.0...v0.0.1) (2026-06-18)


### Features

* add man page generation via cobra/doc ([ffee98e](https://github.com/jamesjohnsdev/issues/commit/ffee98ef54a4c4523ac6541f9dc693b37be2d782))
* added  command ([2d6a124](https://github.com/jamesjohnsdev/issues/commit/2d6a124738a6a9f71f68e5e0d5722fe78f60ea89))
* added `complete` command ([e2fc6a9](https://github.com/jamesjohnsdev/issues/commit/e2fc6a9f9e91f8adce38a7b951249b3550346246))
* **create:** add stepwise interactive prompts when no args given ([3f9a4d4](https://github.com/jamesjohnsdev/issues/commit/3f9a4d480d6436f1539e77f7ffd046339d27c32a))
* delete command ([5a9b7a1](https://github.com/jamesjohnsdev/issues/commit/5a9b7a1f522c1431194ba9f8857e90cdc58ee113))
* enabled creation direct passthrough to editor ([c898d63](https://github.com/jamesjohnsdev/issues/commit/c898d63efffb8d91c805147ce1773f079e588f65))
* implement issues CLI with local/GitHub sync ([6b17f50](https://github.com/jamesjohnsdev/issues/commit/6b17f5025cd57b2c41e073bef135f8da5003c373))
* improved decoration on outputs ([8937aaf](https://github.com/jamesjohnsdev/issues/commit/8937aaf49045890fc72ecd7b71908b21089e7de3))
* improved display for issue list ([f9c8831](https://github.com/jamesjohnsdev/issues/commit/f9c883119704f3e9b8e710ca7b0ebe473b9757d4))
* list command defaults to show open issues only ([a4d3a3c](https://github.com/jamesjohnsdev/issues/commit/a4d3a3c3492bc1acffde5e1eefece14205d3a327))


### Bug Fixes

* correct parsing of local issue numbers ([6d3a9a1](https://github.com/jamesjohnsdev/issues/commit/6d3a9a121161388e5bf7fa2c2ffaed5f46a6b688))
* **create:** three editor mode correctness bugs ([c40a51c](https://github.com/jamesjohnsdev/issues/commit/c40a51c011ef7f3f5d194343b08a7f9c413295f9))
* **root:** rename root command name to "issues" ([4c9c2b8](https://github.com/jamesjohnsdev/issues/commit/4c9c2b85ff40e3474696511be8c39c9888bddd1d))
* **util:** reject partial numeric ids such as "123abc" ([a3f7cf7](https://github.com/jamesjohnsdev/issues/commit/a3f7cf7960a1247346f4a6d40db84b2e8c351ee8))
* **view:** split editor string to support flags ([97ff072](https://github.com/jamesjohnsdev/issues/commit/97ff072c9f8faf51753ef36c9f638f3126ad0f55))
