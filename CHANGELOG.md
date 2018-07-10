# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- A linter to verify that no enum uses the option `allow_alias.`
### Changed
- The protoc-commands command is now accessible via the `dry-run`
  flag for compile, format, gen, and lint.
- The grpc command now takes flags for address, method, and data or stdin
  as opposed to parsing these from variable-length command args.
- If more than one `prototool.yaml` is found for the input directory or files,
  an error is returned.


## [0.4.0] - 2018-06-22
### Added
- A new command `prototool create` to auto-generate Protobuf files from a
  template. The generated files have the Protobuf package, `go_package`,
  `java_multiple_files`, `java_outer_classname`, and `java_package` values set
  depending on the location of your file and config settings. Make sure to
  update your Vim plugin setup as well if using the Vim integration. See the
  documentation for `prototool create` in the README.md for more details.
### Changed
- The values for `java_multiple_files`, `java_outer_classname`, and
  `java_package` that pass lint by default now reflect what is expected
  by the Google Cloud APIs file structure. See
  https://cloud.google.com/apis/design/file_structure for more details.
- `protobuf format` will now automatically update the value of `go_package`,
  `java_multiple_files`, `java_outer_classname`, and `java_package` to match
  what is expected in the default Style Guide. This functionality can be
  suppressed with the flag `--no-rewrite`. See the documentation for
  `prototool format` in the README.md for more details.
- Formatting configuration options are removed. We think there should be
  only one way to format, so we went with defaults of two spaces for indents,
  semicolons at the end of RPCs if there are no RPC options, and always
  having a newline at the end of a file.


## [0.3.0] - 2018-06-14
### Added
- Linters to verify that `java_multiple_files` and `java_outer_classname` are
  unset.
### Fixed
- The formatting order now reflects
  https://cloud.google.com/apis/design/file_structure by moving the location
  of imports to be below syntax, package, and file options.
- Temporary files used for `FileDescriptorSets` are now properly cleaned up.
- Packages that begin with a keyword no longer produce an error when using
  `prototool format` or `prototool lint`.


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

[Unreleased]: https://github.com/uber/prototool/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/uber/prototool/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/uber/prototool/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/uber/prototool/compare/v0.1.0...v0.2.0
