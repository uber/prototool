# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2018-05-29
### Added
- A default lint rule to verify that a package is always declared.
- A lint group `all` that contains all the lint rules, not just the default
  lint rules.
- A flag `--harbormaster` that will print failures in JSON that is compatible
  with the Harbormaster API.

### Fixed
- `prototool init` will return an error if there is an existing prototool.yaml
  file instead of overwriting it.
- Nested options are now properly printed out from `prototool format`.
- Repeated options are now properly printed out from `prototool format`.
- Weak and public imports are now properly printed out from `prototool format`.
- Option keys with empty values are no longer printed out
  from `prototool format`.


## 0.1.0 - 2018-04-11
### Added
- Initial release.

[0.2.0]: https://github.com/uber/prototool/compare/v0.1.0...v0.2.0
