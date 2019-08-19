# Breaking Change Detector

[Back to README.md](README.md)

  * [Usage](#usage)
    * [Git](#git)
    * [Saved State](#saved-state)
  * [Beta vs\. Stable Packages](#beta-vs-stable-packages)
  * [Per\-package breaking change detection](#per-package-breaking-change-detection)
  * [Future source code location references](#future-source-code-location-references)
  * [Implementation](#implementation)

Protobuf is a great way to represent your APIs and generate stubs in each language you develop
with. As such, Protobuf APIs should be stable so as not to break consumers across repositories.
Even in a monorepo context, making sure that your Protobuf APIs do not introduce breaking
changes is important so that different deployed versions of your services do not have
wire incompatibilities.

Prototool exposes a breaking change detector through the `prototool break check` command. This will
check your current Protobuf definitions against a past version of your Protobuf definitions to see
if there are any source or wire incompatible changes. Some notes on this command:

- The breaking change detection operates on a **per-package** basis, not per-file - definitions
  can be moved between files within the same Protobuf package without being considered breaking.
- The breaking change detector can either check against a given git branch or tag, or it can check
  against a previous state saved with the `prototool break descriptor-set` command.
- The breaking change detector understands the concept of **beta vs. stable packages**, discussed
  in the [V2 Style Guide](../style/README.md#package-versioning). By default, the breaking change
  detector will not check beta packages for breaking changes, and will not allow stable packages to
  depend on beta packages, however both of these options are configurable in your `prototool.yaml`
  file.

The following is the list of changes understood to be breaking:

- Deleting or renaming a package.
- Deleting or renaming an enum, enum value, message, message field, service, or service method.
- Changing the type of a message field.
- Changing the tag of a message field.
- Changing the label of a message field, i.e. optional, repeated, required.
- Moving a message field currently in a oneof out of the oneof.
- Moving a message field currently not in a oneof into the oneof.
- Changing the function signature of a method.
- Changing the stream value of a method request or response.

Over the next few minor versions of Prototool, we may add to this list if other changes are
understood to be breaking, for example we may detect file option changes. However, the addition of
additional checks is highly unlikely, and we will update the documentation accordingly when we are
sure that no additional checks will be added.

A note on reserved fields: `prototool break check` expects you to deprecate fields instead of
removing them and adding them to `reserved`. See the
[V2 Style Guide's discussion](../style/README.md#reserved-keyword) on the `reserved` keyword for
more details.

## Usage

### Git

The breaking change detector can either be used to check definitions against a commit on a git
branch or tag, or can check against a previously saved state. The only non-flag input argument
is the directory to check, similar to other `prototool` commands.

Assuming your Protobuf definitions are in `path/to/proto` (which is usually where your
`prototool.yaml` file will be as well):

```bash
# Checks against the default git branch of your repository
prototool break check path/to/proto
# Checks against the git branch or tag "dev"
prototool break check path/to/proto --git-branch dev
```

What this does behind the scenes:

- Compiles all your Protobuf definitions in `path/to/proto` into a `FileDescriptorSet`.
- Calls `git clone --depth 1`, optionally with `--branch` if the `--git-branch` flag was specified,
  and makes a clone of the git repository at your current directory into a temporary directory.
- Compiles all the Protobuf definitions in `path/to/proto` in this temporary clone into a
  `FileDescriptorSet`.
- Compares the two `FileDescriptorSets` to see if breaking changes were introduced from the current
  state to the previous state.
- Deletes the temporary clone.

For this to work, you must run `prototool` from the root of a `git` checkout, and the directory
path given to `prototool` must be relative. You can also specify no directory, as with other
`prototool` commands, and `prototool` will assume your current directory contains all Protobuf
definitions.

### Saved State

If not using git, or you would prefer not to rely on git state to compare your Protobuf definitions
for breaking changes, you can instead save state to a file and use that for comparsion.

To save your current state:

```bash
prototool break descriptor-set path/to/proto -o break_descriptor_set.bin
```

This is in effect an alias for `prototool descriptor-set --include-imports -o FILE`, however the
`-o` flag is required.

To check your current definitions against a previous state:

```bash
prototool break check path/to/proto --descriptor-set-path break_descriptor_set.bin
```

What this does behind the scenes:

- Compiles all your Protobuf definitions in `path/to/proto` into a `FileDescriptorSet`.
- Deserializes `break_descriptor_set.bin` into a `FileDescriptorSet`.
- Compares the two `FileDescriptorSets` to see if breaking changes were introduced from the current
  state to the previous state.

## Beta vs. Stable Packages

As described in the [V2 Style Guide](../style/README.md#package-versioning), `prototool`
understands the concept of beta vs. stable packages.

If a package's last component is `vMAJORbetaBETA`, where `MAJOR` and `BETA`
are both greater than 0, `prototool break check` will understand that this package is a
beta package, otherwise the package is understood as a stable pacakge.

The following are examples of beta packages.

```proto
package uber.trip.v1beta1;
package uber.user.v1beta2;
package uber.road.v2beta1;
```

By default, Prototool will not check beta packages for breaking changes, and will also
check to make sure not stable packages depend on beta packages. Both of these options
are configurable in your `prototool.yaml`.

```yaml
break:
  # Include beta packages in breaking change detection.
  include_beta: true
  # Allow stable packages to depend on beta packages.
  # If include_beta is true, this is implicitly set.
  allow_beta_deps: true
```

## Per-package breaking change detection

Some notes on why we chose per-package instead of per-file logic for breaking change detection.

Moving definitions between files within the same Protobuf package does not result in a breaking
change on either a source code or wire level. For example, one can move `message Foo` within
package `bar.baz.v1` from file `bat.proto` to `bam.proto` without any effect on the resulting
source code or over-the-wire representation (with a few minor exceptions, see below).

Refactoring file structure isn't just common, it's encouraged in many cases. We've seen a lot of
instances of files being named in inconsistent manners, and instances where types logically belong
in other files. If we were to enforce breaking changes on a per-file basis, this would encourage
inconsistent file structure as schemas evolve, with no benefit (since there is no actual breaking
change). We want breaking change detection to prevent actual breakages in the API, but we don't
want to limit developers beyond that as it will have the effect of producing APIs that are even
more inconsistent, which is against Prototool's main goal.

There are two main exceptions to the "no breaking changes in generated code" statement:

- In Java, we typically have the `java_outer_classname` file options match the Protobuf file name.
  This means that types generated inside this classname would change for Java. However, with
  `java_multiple_files = true` set (also standard, and largely the standard across the Protobuf
  community including the Google Cloud APIs), there are only two public methods that go inside this
  class: `registerAllExtensions` and `getDescriptor`. Both of these methods should not be manually
  relied on outside of generated code, and it's such an edge case/advanced case that it's not worth
  enforcing per-file breaking changes given the drawbacks.
- In C++ and Python (and other languages that rely on per-file imports), the file imports may need
  to change. For example, importing "bat.pb.h" may need to change to "bam.pb.h". However, since
  this is something that can be detected at compile-time (and should be detectable at testing time
  for dynamic languages), and given the drawbacks of per-file breaking change detection, we still
  prefer per-package breaking change detection.

## Future source code location references

For now, unlike other `prototool` commands, `prototool break check` only outputs error messages
and **does not output filename:line:column location references.** The problem of referencing your
current Protobuf files with the location of a breaking change is harder than it seems, however we
aim to implement this in the future. The logic that will likely be implemented is as follows.

- For deleted enum values, message fields, or service methods, point to the encapsulating enum,
  message, or service.
- For renamed fields, fields with a type change or tag change, or service methods whose signature
  changes, point to the field or service method.
- For deleted enums, messages, or services, point to `1:1` of the file where this was deleted.
  Determining the file to put this on will be not technically possible if the file no longer exists
  or is renamed, so in this case, either do not output a filename and default to `<input>`, or
  choose the first file in the package alphabetically if a file is required. If all files were
  deleted, `<input>` will likely have to be defaulted to.

## Implementation

The breaking change detection is primarily implemented through a series of packages within
`internal`.

- [internal/reflect](../internal/reflect) - This wraps the Protobuf definition at
  [internal/reflect/proto/uber/proto/reflect/v1/reflect.proto](../internal/reflect/proto/uber/proto/reflect/v1/reflect.proto),
  which is intended to contain definitions that represent Protobuf definitions on a per-package
  basis. This file is intended to be analogous to the [descriptor.proto](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/descriptor.proto)
  file in Protobuf's source code, however, as of now, `reflect.proto` only contains the information
  necessary for breaking change detection. This could be extended in the future and exposed outside
  of Prototool's internal implementation.
- [internal/extract](../internal/extract) - This wraps `internal/reflect` in a manner that makes
  working with the `reflect.proto` definitions easier within Golang.
- [internal/breaking](../internal/breaking) - This uses `internal/extract` to implement the actual
  breaking change logic.
- [internal/git](../internal/git) - This provides the function to do temporary git clones.
