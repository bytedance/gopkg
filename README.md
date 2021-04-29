# gopkg

`gopkg` is a universal utility collection for Go, it complements offerings such as Boost, Better std, Cloud tools.

## Table of Contents

- [Introduction](#Introduction)
- [Catalogs](#Catalogs)
- [Releases](#Releases)
- [License](#License)

## Introduction

`gopkg` is a universal utility collection for Go, it complements offerings such as Boost, Better std, Cloud tools. It is migrated from the internal code base at ByteDance and has been extensively adopted in production.

We depend on the same code(this repo) in our production environment.

## Catalogs

* [cache](https://github.com/bytedance/gopkg/tree/main/cache): Caching Mechanism
* [cloud](https://github.com/bytedance/gopkg/tree/main/cloud): Cloud Computing Design Patterns
* [collection](https://github.com/bytedance/gopkg/tree/main/collection): Data Structures
* [lang](https://github.com/bytedance/gopkg/tree/main/lang): Enhanced Standard Libraries
* [util](https://github.com/bytedance/gopkg/tree/main/util): Utilities Useful across Domains

## Releases

`gopkg` recommends users to "live-at-head" (update to the latest commit from the main branch as often as possible).
We develop at `develop` branch and will only merge to `main` when `develop` is stable.

## License

`gopkg` is licensed under the terms of the Apache license 2.0. See [LICENSE](LICENSE) for more information.
