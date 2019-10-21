# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]
- No changes yet.


## [1.9.0] - 2019-10-12
- Updated dependencies for Go 1.13


## [1.8.0] - 2019-06-10
- Update the default `protoc` version to `3.8.0`.
- Parse updated `protoc` output for `3.8.0`.
- Fix issue where there were an unbounded number of `protoc` calls were
  being executed.


## [1.7.0] - 2019-05-02
- Fix issue where `config init --document` produced an invalid YAML file.
- Dependency updates.


## [1.6.0] - 2019-04-05
- Dynamically resolve `google.protobuf.Any` values for gRPC error details.


## [1.5.0] - 2019-04-03
- Add linters for enum field and message field comments. These linters are not
  part of any lint group but can be manually added in a configuration file.
- Add `--generate-ignores` flag to the `lint` command to print out the value
  for `lint.ignores` that will allow `lint` to pass. This is useful when
  migrating to a set of lint rules, usually a lint group.
- Update the default version of `protoc` to `3.7.1`.


## [1.4.0] - 2019-03-19
- Add concept of lint groups. The default lint group is named `uber1`. The lint
  group can be specified with the `lint.group` option.
- New `uber2` lint group and associated V2 Style Guide representing the second
  version of our lint rules. These rules are almost entirely a superset of the
  V1 Style guide lint rules. If `lint.group` is set to `uber2`, this also will
  affect the `create` and `format` commands, as the `uber2` lint group adds
  more file options to more closely match the [Google Cloud APIs File Structure](https://cloud.google.com/apis/design/file_structure)
  and changes the value of `go_package` to take versions into account.
  In total, 39 lint rules have been added as compared to the `uber1` lint
  group.
- New `google` lint group representing Google's minimal [Style Guide](https://developers.google.com/protocol-buffers/docs/style).
- Add `--list-lint-group` flag to the `lint` command to list a lint group's
  rules.
- Add `--diff-lint-groups` flag to the `lint` command to print the diff
  between two lint groups.
- Add `descriptor-set` command to output a merged `FileDescriptorSet`
  with all files compiled to either stdout, a given file, or a temporary file.
  Useful with external tools that use FileDescriptorSets, and also useful for
  inspection if the `--json` flag is given.
- Add breaking change detector as the `break check` command. By default, this
  compiles your existing Protobuf definitions, and then does a shallow clone
  of your git repository against the default branch and compiles the
  definitions on that branch, and compares the existing versus the branch.
  The branch can be controlled with the `--git-branch` flag, and one can
  use a `FileDescriptorSet` instead of a shallow clone by generating a
  file with `break descriptor-set` and then passing the path to this file
  to `break check` with the `--descriptor-set-path` flag.
- A Docker image is now provided on Docker Hub as [uber/prototool](https://hub.docker.com/r/uber/prototool)
  which provides an environment with commonly-used plugins.
- Switch to Golang Modules for dependency management.
- Add Bazel build files and `bazel/deps.bzl` to allow Prototool to be easily
  built within a Bazel workspace.
- Add `lint.file_header` option to allow a file header to be specified. This
  affects `lint`, `format`, and `create`.
- Allow `generate.plugins.path` to be relative. If a relative path is given,
  Prototool will search your `PATH` for the specified executable.
- Add `generate.plugins.file_suffix` option that allows for JAR generation with
  the built-in `java` plugin, and `FileDescriptorSet` generation with the
  built-in `descriptor_set` plugin.
- Add `generate.plugins.include_imports` and
  `generate.plugins.include_source_info` to be used with the built-in
  `descriptor_set` plugin.
- Add `cache` top-level command to allow management of the `protoc` cache.
- Add `x` top-level command for experimental functionality.
- Add `inspect` command under `x` with Protobuf inspection capabilities.
- Add `--error-format` flag to allow specific error fields to be printed.
- Allow the `protoc` binary and WKT paths to be controlled by the environment
  variables `PROTOTOOL_PROTOC_BIN_PATH` and `PROTOTOOL_PROTOC_WKT_PATH` in
  addition to the existing `--protoc-bin-path` and `--protoc-wkt-path` flags.
  The flags take precedence. This is especially useful for Docker images.
- Add file locking around the `protoc` downloader to eliminate concurrency
  issues where multiple `prototool` invocations may be accessing the cache
  at the same time.
- Add TLS support to the `grpc` command.
- Add `--details` flag to the `grpc` command to output headers, trailers,
  and statuses as well as the responses.
- Unix domain sockets can now be specified for the `--address` flag of the
  `grpc` command via the prefix `unix://`.


## [1.3.0] - 2018-09-17
### Added
- Accept `prototool.json` files for configuation in addition to
  `prototool.yaml` files.
- Add `--config-data` flag.
- Add `--protoc-bin-path` and `--protoc-wkt-path` flags to manually
  set the paths for where `protoc` is run and where the
  Well-Known Types are included from.


## [1.2.0] - 2018-08-29
### Added
- Add `json` flag to `all`, `compile`, `format`, `generate` and `lint` commands.


## [1.1.0] - 2018-08-24
### Added
- Add support for Homebrew builds.


## [1.0.0] - 2018-08-23
- Initial release.


## [1.0.0-rc1] 2018-08-16
### Fixed
- Fixed regression where `prototool version` did not output 'Git commit' and
  'Built'.


## [0.7.1] 2018-08-15
### Fixed
- Fixed an issue where Golang `Mname=package` modifiers were being duplicated.


## [0.7.0] - 2018-08-09
### Changed
- Move `protoc_includes` and `protoc_version` settings under `protoc` key.
- Move `allow_unused_imports` to `protoc.allow_unused_imports`.
- Move `protoc-url` global flag under the applicable commands: all,
  compile, format, gen, and lint.
- Rename `gen` to `generate`.


## [0.6.0] - 2018-08-03
### Changed
- Delete the ability to explicitly specify multiple files, and have the effect
  of one file being specified be the same as the former `--dir-mode`. See
  [#16](https://github.com/uber/prototool/issues/16) for more details.
- Delete `protoc_include_wkt` setting. This is always set to true.
- Delete `no_default_excludes` setting. This is always set to true.
- Delete `gen.go_options.no_default_modifiers` setting.
- Delete `lint.group` setting.
- Delete `harbormaster` global flag.
- Refactor `create.dir_to_base_package` to the list `create.packages` See
  the documentation for more details.
- Rename `create.dir_to_base_package` -> `create.dir_to_package`.
- Move `prototool init` to `prototool config init`.
- Move `gen.plugin_overrides` to `gen.plugins.path`.
- Refactor `lint` configuration. See the documentation for details.
- Refactor `format --no-rewrite` so that the previous default is now enabled via
  `format --fix`.
### Fixed
- Fix `excludes` setting to correctly match file path prefixes.


## [0.5.0] - 2018-07-26
### Added
- A linter to verify that no enum uses the option `allow_alias.`
- The `--protoc-url` flag can now handle references to local protoc zip files
  as well as normal http references by handling urls of the form
  `file:///path/to/protoc.zip`.
### Changed
- The formatter now prints primitive field options on the same line
  as the field.
- The commands `binary-to-json`, `clean`, `descriptor-proto`, `download`,
  `field-descriptor-proto`, `json-to-binary`, `list-all-linters`,
  `list-all-lint-groups`, `list-linters`, `list-lint-group`, and
  `service-descriptor-proto` are deleted to reduce the surface area
  for the v1.0 release.
- The commands `list-all-linters` and `list-linters` are now flags
  on the `lint` command.
- The flags `--cache-path` and `--print-fields` are deleted to reduce the
  surface area for the v1.0 release.
- The option `lint.group` in the `prototool.yaml` configuration is deleted
  to reduce the surface area for the v1.0 release.
- The command `protoc-commands` is now accessible via the `--dry-run`
  flag on the commands `compile` and `gen`.
- The `grpc` command now takes the flags `--address`, `--method`, and `--data`
  or `--stdin` as opposed to parsing these from variable-length command args.
- If more than one `prototool.yaml` is found for the input directory or files,
  an error is returned.
- The `prototool` binary package is moved under `internal`.


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

[Unreleased]: https://github.com/uber/prototool/compare/v1.9.0...HEAD
[1.9.0]: https://github.com/uber/prototool/compare/v1.8.0...v1.9.0
[1.8.0]: https://github.com/uber/prototool/compare/v1.7.0...v1.8.0
[1.7.0]: https://github.com/uber/prototool/compare/v1.6.0...v1.7.0
[1.6.0]: https://github.com/uber/prototool/compare/v1.5.0...v1.6.0
[1.5.0]: https://github.com/uber/prototool/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/uber/prototool/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/uber/prototool/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/uber/prototool/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/uber/prototool/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/uber/prototool/compare/v1.0.0-rc1...v1.0.0
[1.0.0-rc1]: https://github.com/uber/prototool/compare/v0.7.1...v1.0.0-rc1
[0.7.1]: https://github.com/uber/prototool/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/uber/prototool/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/uber/prototool/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/uber/prototool/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/uber/prototool/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/uber/prototool/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/uber/prototool/compare/v0.1.0...v0.2.0
