# Changelog

## [1.0.1](https://github.com/DarkoKlisuric/godebugbar/compare/v1.0.0...v1.0.1) (2026-11-14)


### Bug Fixes

* keep the toolbar on the latest request and the current page ([7fe2d88](https://github.com/DarkoKlisuric/godebugbar/commit/7fe2d8867a1efdfc12e17d8d6bd57153dd181ff5))
* keep the toolbar on the latest request and the current page ([fbd2122](https://github.com/DarkoKlisuric/godebugbar/commit/fbd21221a56d47ecd40b85e5485f2f43db1f2476))

## [1.0.0](https://github.com/DarkoKlisuric/godebugbar/compare/v0.4.0...v1.0.0) (2026-11-14)


### ⚠ BREAKING CHANGES

* profiling is disabled until WithEnabled(true) is passed. Option is now func(*engine.Engine), so only this module can create options. WithCORS ignores "*". WithExcludedMethods adds to OPTIONS instead of replacing it. GET /_debugbar/api/profiles returns summaries; fetch /api/profiles/{id} for the full profile. The request, runtime, log, HTTP client and timeline collectors are registered by default.

### Features

* add pgx and database/sql adapters ([11558cf](https://github.com/DarkoKlisuric/godebugbar/commit/11558cfbd101f6bb73b8df54dc36a4a67f432747))
* harden the profiler and add a public recording API ([ade1410](https://github.com/DarkoKlisuric/godebugbar/commit/ade1410bea24a13f405d0bc189dbd8e66e2fbf17))
* **ui:** rewrite the toolbar and profiler page ([8a3769b](https://github.com/DarkoKlisuric/godebugbar/commit/8a3769bcff25f359688edc54d17bac2bff0c9d1f))

## [0.4.0](https://github.com/DarkoKlisuric/godebugbar/compare/v0.3.1...v0.4.0) (2026-11-14)


### Features

* add standalone profiler dashboard, CORS support, and opt-in collector options ([1be9b88](https://github.com/DarkoKlisuric/godebugbar/commit/1be9b88dd83e5fcdf0eee91ff779d76d784cdcad))
* add standalone profiler dashboard, CORS support, and opt-in collector options ([245a756](https://github.com/DarkoKlisuric/godebugbar/commit/245a756cacb854b0a0a1931d085b5ca0215ba815))

## [0.3.1](https://github.com/DarkoKlisuric/godebugbar/compare/v0.3.0...v0.3.1) (2026-11-14)


### Code Refactoring

* encapsulate internal implementation under internal packages ([721a5b0](https://github.com/DarkoKlisuric/godebugbar/commit/721a5b0a6f3ff5a7740e3eed68e7355faa986929))
* encapsulate internal implementation under internal packages ([94b53f1](https://github.com/DarkoKlisuric/godebugbar/commit/94b53f14cf85d4b8d6576253a0c2a52128c8b614))

## [0.3.0](https://github.com/DarkoKlisuric/godebugbar/compare/v0.2.0...v0.3.0) (2026-11-14)


### Features

* add default collectors, gin and gorm adapters ([74d3277](https://github.com/DarkoKlisuric/godebugbar/commit/74d327705c5f6c9a26b6f2330c112e1adca93b63))
* add default request, query and runtime collectors ([ea42995](https://github.com/DarkoKlisuric/godebugbar/commit/ea429954124b04f5dd12f7f3e6efac45da23c3e6))
* add the gin and gorm adapters ([62b34af](https://github.com/DarkoKlisuric/godebugbar/commit/62b34afe52f888d0f00b18ff0cd531601e9f8cfd))

## [0.2.0](https://github.com/DarkoKlisuric/godebugbar/compare/v0.1.0...v0.2.0) (2026-11-14)


### Features

* add the collector, panel and profile contracts ([584ce19](https://github.com/DarkoKlisuric/godebugbar/commit/584ce197d62e16a725a4321d51686e689c7ee7f4))
* add the core profiler and collector contracts ([1f306e4](https://github.com/DarkoKlisuric/godebugbar/commit/1f306e47e24d83d47b6b1640e6cee8a4948ca3d1))
* add the core profiler and http middleware ([5f5fd71](https://github.com/DarkoKlisuric/godebugbar/commit/5f5fd7127815dcd295bf283562e06daf9d635130))

## Changelog
