# Uber Protobuf Style Guide V2

  * [V1 versus V2 Style Guides](#v1-versus-v2-style-guides)
  * [Spacing](#spacing)
  * [Package Naming](#package-naming)
  * [Package Versioning](#package-versioning)
  * [Directory Structure](#directory-structure)
  * [File Structure](#file-structure)
  * [Syntax](#syntax)
  * [File Options](#file-options)
  * [Imports](#imports)
  * [Reserved Keyword](#reserved-keyword)
  * [Enums](#enums)
    * [Enum Value Names](#enum-value-names)
    * [Nested Enums](#nested-enums)
  * [Messages](#messages)
    * [Message Fields](#message-fields)
    * [Oneofs](#oneofs)
    * [Nested Messages](#nested-messages)
  * [Services/RPCs](#servicesrpcs)
    * [Streaming RPCs](#streaming-rpcs)
    * [HTTP Annotations](#http-annotations)
  * [Naming](#naming)
  * [Documentation](#documentation)

This is the V2 of the Uber Protobuf Style Guide.

See the [uber](uber) directory for an example of all concepts explained in this Style Guide.

## V1 versus V2 Style Guides

The V2 Style Guide contains 39 new lint rules compared to our [V1 Style Guide](../etc/style/uber1/uber1.proto)
that represent lessons we have learned in our API development. The rules within this Style Guide
will help you write clean, consistent, and maintainable APIs. We recommend following the V2 Style
Guide instead of the V1 Style Guide - once you understand the rules described below, and are used
to developing with them, it becomes simple to follow.

However, for backwards-compatibility reasons, the V1 Style Guide is the default style guide
enforced by Prototool if no lint group configuration is present. To use the V2 Style Guide with
Prototool, set the following in your `prototool.yaml`.

```yaml
lint:
  group: uber2
```

The V2 Style Guide is almost entirely a superset of the V1 Style Guide. There are only two items
that are incompatible.

- Nested enum value names do not need to have their message type as part of their prefix.
- The `go_package` file option is now suffixed by the package version and not `pb`.

We call out these differences below. Even if you are primarily using the V1 Style Guide, we still
recommend you follow the remainder of the rules in the V2 Style Guide, although they will not
be enforced by Prototool unless the `uber2` lint group is set.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Spacing

Use two spaces instead of tabs. There are many schools of thought here, this is the one we chose.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Package Naming

Some conventions:

- A **package name** is a full package name, i.e. `uber.trip.v1`.
- A **package sub-name** is a part of a package name, ie `uber`, `trip`, or `v1`.
- A **package version** is the last package sub-name that specifies the version,
  i.e. `v1`, `v1beta1`, or `v2`.

Package sub-names should be short and descriptive, and can use abbreviations if necessary.
Package sub-names should only include characters in the range `[a-z0-9]`, i.e always lowercase
and with only letter and digit characters. If names get too long or have underscores, the
generated stubs in certain languages are less than idiomatic.

As illustrative examples, the following are not acceptable package names.

```proto
// Examples of bad package names.

// Note that specifying multiple packages is not valid Protobuf, however
// we do this here for brevity.

// The package sub-name credit_card_analysis is not short, and contains underscores.
package uber.finance.credit_card_analysis.v1;
// The package sub-name creditcardanalysisprocessing is longer than desired.
package uber.finance.creditcardanalysisprocessing.v1;
```

The following are acceptable package names.

```proto
// Each package sub-name is short and to the point.
package uber.trip.v1;
// Grouping by finance and then payment is acceptable.
package uber.finance.payment.v1;
// Ccap is for credit card analysis processing.
package uber.finance.ccap.v1;
```

Package sub-names cannot be any of the following.

- `internal` - This is effectively a reserved keyword in Golang and results in the generated
  package not being accessible outside of it's context.
- `public` - This is a reserved keyword in many languages.
- `private` - This is a reserved keyword in many languages.
- `protected` - This is a reserved keyword in many languages.
- `std` - While not a reserved keyword in C++, this results in generated C++ stubs that do not
  compile.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Package Versioning

The last package sub-name should be a major version of the package, or the major version
followed by the beta version, specified as `vMAJOR` or `vMAJORbetaBETA`, where `MAJOR` and `BETA`
are both greater than 0. The following are examples of acceptable package names.

```proto
package uber.trip.v1beta1;
package uber.trip.v1beta2;
package uber.trip.v1;
package uber.trip.v2beta1;
package uber.trip.v2;
package something.v2;
```

As illustrative examples, the following are not acceptable package names.

```proto
// No version.
package uber.trip;
// Major version is not greater than 0.
package uber.trip.v0;
// Beta version is not greater than 0.
package uber.trip.v1beta0;
```

Packages with only a major version are considered **stable** packages, and packages with a major
and beta version are considered **beta** packages.

Breaking changes should never be made in stable packages, and stable packages should never depend
on beta packages. Both wire-incompatible and source-code-incompatible changes are considered
breaking changes. The following are the list of changes currently understood to be breaking.

- Deleting or renaming a package.
- Deleting or renaming an enum, enum value, message, message field, service, or service method.
- Changing the type of a message field.
- Changing the tag of a message field.
- Changing the label of a message field, i.e. optional, repeated, required.
- Moving a message field currently in a oneof out of the oneof.
- Moving a message field currently not in a oneof into the oneof.
- Changing the function signature of a method.
- Changing the stream value of a method request or response.

Beta packages should be used with extreme caution, and are not recommended.

Instead of making a breaking change, rely on deprecation of types.

```proto
// Note that all enums, messages, services, and service methods require
// sentence comments, and each service must be in a separate file, as
// outlined below, however we omit this here for brevity.

enum Foo {
  option deprecated = true;
  FOO_INVALID = 0;
  FOO_ONE = 1;
}

enum Bar {
  BAR_INVALID = 0;
  BAR_ONE = 1 [deprecated = true];
  BAR_TWO = 2;
}

message Baz {
  option deprecated = true;
  int64 one = 1;
}

message Bat {
  int64 one = 1 [deprecated = true];
  int64 two = 2;
}

service BamAPI {
  option deprecated = true;
  rpc Hello(HelloRequest) returns (HelloResponse) {}
}

service BanAPI {
  rpc SayGoodbye(SayGoodbyeRequest) returns (SayGoodbyeResponse) {
    option deprecated = true;
  }
}
```

If you really want to make a breaking change, or just want to clean up a package, make a new
version of the package by incrementing the major version and copy your definitions as
necessary. For example, copy `foo.bar.v1` to `foo.bar.v2`, and do any cleanups required.
This is not a breaking change as `foo.bar.v2` is a new package. Of course, you are responsible
for the migration of your callers.

Prototool is able to detect breaking changes in your Protobuf definitions. See
[breaking.md](../docs/breaking.md) for more details.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Directory Structure

Files should be stored in a directory structure that matches their package sub-names. All files
in a given directory should be in the same package.

The following is an example of this in practice.

```
.
└── uber
    ├── finance
    │   ├── ccap
    │   │   └── v1
    │   │       └── ccap.proto // package uber.finance.ccap.v1
    │   └── payment
    │       ├── v1
    │       │   └── payment.proto // package uber.payment.v1
    │       └── v1beta1
    │           └── payment.proto // package uber.payment.v1beta1
    └── trip
        ├── v1
        │   ├── trip_api.proto // package uber.trip.v1
        │   └── trip.proto // package uber.trip.v1
        └── v2
            ├── trip_api.proto // package uber.trip.v2
            └── trip.proto // package uber.trip.v2
```

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## File Structure

Files should be named `lower_snake_case.proto`.

All files should be ordered in the following manner.

   1. License Header (if applicable)
   2. Syntax
   3. Package
   4. File options (alphabetized)
   5. Imports (alphabetized)
   6. Everything Else

Protobuf definitions should go into one of two types of files: **Service files** or
**Supporting files**.

A service file contains exactly one service, and it's corresponding request and response messages.
This file is named after the service, substituting `PascalCase` for `lower_snake_case`. The service
should be the first element in the file, with requests and responses sorted to match the order
of the declared service methods.

A supporting file contains everything else, i.e. enums, and messages that are not request or
response messages. These files have no enforced naming structure or otherwise, however the general
recommendation is that if you have less than 15 definitions, they should all go in a file named
after the last non-version package sub-name. For example, for a package `uber.trip.v1` with less
than 15 non-service-related definitions, you would have a single supporting file
`uber/trip/v1/trip.proto`. While there are arguments for and against the single file
recommendation, this provides the easiest mechanism to normalize file structure across a repository
of Protobuf files while making it simple for users to grok a Protobuf package without having to
change between multiple files, each requiring many imports.

The following is an example of a supporting file with two definitions. Note that it is merely
coincidence that the file is named `trip.proto` and it contains a definition `Trip` - there is
no need to name files by a type or types contained within them, other than for services.

```proto
syntax = "proto3";

package uber.trip.v1;

option csharp_namespace = "Uber.Trip.V1";
option go_package = "tripv1";
option java_multiple_files = true;
option java_outer_classname = "TripProto";
option java_package = "com.uber.trip.v1";
option objc_class_prefix = "UTX";
option php_namespace = "Uber\\Trip\\V1";

import "uber/user/v1/user.proto";
import "google/protobuf/timestamp.proto";

// A trip taken by a rider.
message Trip {
  string id = 1;
  user.v1.User user = 2;
  google.protobuf.Timestamp start_time = 3;
  google.protobuf.Timestamp end_time = 4;
  repeated Waypoint waypoints = 5;
}

// A given waypoint.
message Waypoint {
  // In the real world, addresses would be normalized into
  // a PostalAddress message or such, but for brevity we simplify
  // this to a freeform string.
  string postal_address = 1;
  // The nickname of this waypoint, if any.
  string nickname = 2;
}
```

The following is an example of a service file named `uber/trip/v1/trip_api.proto` showing a service
`TripAPI` with two service methods, and requests and responses ordered by method. Note that
request and response messages do not require comments.

```proto
syntax = "proto3";

package uber.trip.v1;

option csharp_namespace = "Uber.Trip.V1";
option go_package = "tripv1";
option java_multiple_files = true;
option java_outer_classname = "TripApiProto";
option java_package = "com.uber.trip.v1";
option objc_class_prefix = "UTX";
option php_namespace = "Uber\\Trip\\V1";

import "uber/trip/v1/trip.proto";
import "google/protobuf/timestamp.proto";

// Handles interaction with trips.
service TripAPI {
  // Get the trip specified by the ID.
  rpc GetTrip(GetTripRequest) returns (GetTripResponse);
  // List the trips for the given user before a given time.
  //
  // If the start index is beyond the end of the available number
  // of trips, an empty list of trips will be returned.
  // If the start index plus the size is beyond the available number
  // of trips, only the number of available trips will be returned.
  rpc ListUserTrips(ListUserTripsRequest) returns (ListUserTripsResponse);
}

message GetTripRequest {
  string id = 1;
}

message GetTripResponse {
  Trip trip = 1;
}

message ListUserTripsRequest {
  string user_id = 1;
  google.protobuf.Timestamp before_time = 2;
  // The start index for pagination.
  uint64 start = 3;
  // The maximum number of trips to return.
  uint64 max_size = 4;
}

message ListUserTripsResponse {
  repeated Trip trips = 1;
  // True if more trips are available.
  bool next = 2;
}
```

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Syntax

The syntax for Protobuf files should always be `proto3`. It is acceptable to import `proto2` files
for legacy purposes, but new definitions should conform to the newer `proto3` standard.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## File Options

File options should be alphabetized. All files should specify a given set of file options that
largely conform to the [Google Cloud APIs File Structure](https://cloud.google.com/apis/design/file_structure).
Note that `prototool create` and `prototool format --fix` automate this for you, and this can be done
as part of your generation pipeline, so there is no need to conform to this manually.

The following are the required file options for a given package `uber.trip.v1` for a file named
`trip_api.proto`.

```proto
syntax = "proto3";

package uber.trip.v1;

// The csharp_namespace should be the package name with each package sub-name capitalized.
option csharp_namespace = "Uber.Trip.V1";
// The go_package should be the last non-version package sub-name concatenated with the
// package version.
//
// Of special note: For the V1 Style Guide, this was the last package sub-name concatenated
// with "pb", so for a package "uber.trip", it would be "trippb", but since the V2 Style Guide
// requires package versions, we have changed "pb" to be the package version.
option go_package = "tripv1";
// The java_multiple_files option should always be true.
option java_multiple_files = true;
// The java_outer_classname should be the PascalCased file name, removing the "." for the
// extension.
option java_outer_classname = "TripApiProto";
// The java_package should be "com." plus the package name.
option java_package = "com.uber.trip.v1";
// The objc_class_prefix should be the uppercase first letter of each package sub-name,
// not including the package-version, with the following rules:
//   - If the resulting abbreviation is 2 characters, add "X".
//   - If the resulting abbreviation is 1 character, add "XX".
//   - If the resulting abbreviation is "GBP", change it to "GPX". "GBP" is reserved
//     by Google for the Protocol Buffers implementation.
option objc_class_prefix = "UTX";
// The php_namespace is the same as the csharp_namespace, with "\\" substituted for ".".
option php_namespace = "Uber\\Trip\\V1";
```

While it is unlikely that a given organization will use all of these file options for their
generated stubs, this provides a universal mechanism for specifying these options that matches
the Google Cloud APIs File Structure, and all of these file options are built in.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Imports

Imports should be alphabetized.

Imports should all start from the same base directory for a given repository, usually the root of
the repository. For local imports, this should match the package name, so if you have a file
`uber/trip/v1/trip.proto` with package `uber.trip.v1`, you should import it as
`uber/trip/v1/trip.proto`. For external imports, this should generally also be the root of the
repository. For example, if importing [googleapis](https://github.com/googleapis/googleapis)
definitions, you would import `google/logging/v2/logging.proto`, not `logging/v2/logging.proto`,
`v2/logging.proto`, and such.

Imports should never be `public` or `weak`.

Note that the
[Well-Known Types](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf)
should be used whenever possible, and imported starting with `google/protobuf`, for example
`google/protobuf/timestamp.proto`. Prototool provides all of these out of the box for you.

```
.
└── google
    └── protobuf
        ├── any.proto
        ├── api.proto
        ├── compiler
        │   └── plugin.proto
        ├── descriptor.proto
        ├── duration.proto
        ├── empty.proto
        ├── field_mask.proto
        ├── source_context.proto
        ├── struct.proto
        ├── timestamp.proto
        ├── type.proto
        └── wrappers.proto
```

These are available for browsing at
[github.com/protocolbuffers/protobuf/src/google/protobuf](https://github.com/protocolbuffers/protobuf/tree/master/src/google/protobuf)
and are also included in the `include` directory of each [Protobuf Releases ZIP
file](https://github.com/protocolbuffers/protobuf/releases).

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Reserved Keyword

Do not use the `reserved` keyword in messages or enums. Instead, rely on the `deprecated` option.

The following is an example of what not to do.

```proto
// A user.
message User {
  // Do not do this.
  reserved 2, 4;
  reserved "first_name", "middle_names"

  string id = 1;
  string last_name = 3;
  repeated string first_names = 4;
}

// The type of the trip.
enum TripType {
  // Do not do this.
  reserved 2;
  reserved "TRIP_TYPE_POOL";

  TRIP_TYPE_INVALID = 0;
  TRIP_TYPE_UBERX = 1;
  TRIP_TYPE_UBER_POOL = 3;
}
```

Instead, do the following.

```proto
// A user.
message User {
  string id = 1;
  string last_name = 3;
  repeated string first_names = 5;

  string first_name = 2 [deprecated = true];
  repeated string middle_names = 4 [deprecated = true];
}

// The type of the trip.
enum TripType {
  TRIP_TYPE_INVALID = 0;
  TRIP_TYPE_UBERX = 1;
  TRIP_TYPE_UBER_POOL = 3;

  TRIP_TYPE_POOL = 2 [deprecated = true];
}
```

By far the most important reason to use `deprecated` instead of `reserved` is that while not a wire
breaking change, deleting a field or enum value is a source code breaking change. This will result
in code that does not compile, which we take a strong stance against, including in Prototool's
breaking change detector. Your API as a whole should not need semantic versioning - one of the core
promises of Protobuf is forwards and backwards compatibility, and this should extend to your
code as well. This is especially true if your Protobuf definitions are used across multiple
repositories, and even if you have a single monorepo for your entire organization, any external
customers who use your Protobuf definitions should not be broken either.

As of proto3, JSON is also a first-class citizen with respect to Protobuf, and JSON serialization
of Protobuf types uses field names, not tags. This means that if you have JSON consumers, re-using
the same field name is a breaking change. Therefore, when removing a field, if you ever use JSON
or ever could use JSON, you must reserve both the tag and the field name, as done in the above
bad example. If you have to reserve these, it is easier and cleaner to just deprecate the field.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Enums

Enums should always be `PascalCase`. Enum values should be `UPPER_SNAKE_CASE`. The enum option
`allow_aliases` should never be used.

All enums require a comment that contains at least one complete sentence. See the
[Documentation](#documentation) section for more details.

There are many cases when it's tempting to use a string or integer to represent a value that has a
small, finite, and relatively static number of values. These values should almost always be
represented as enums and not strings or integers. An enum value carries semantic meaning and there
is no ability for incorrect values to be set.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Enum Value Names

Enum values have strict naming requirements.

  1. All enum values must have the name of the enum prefixed to all values as `UPPER_SNAKE_CASE`.

For example, for an enum `TripType`, all values must be prefixed with `TRIP_TYPE_`.

```proto
// The type of the trip.
enum TripType {
  TRIP_TYPE_INVALID = 0;
  TRIP_TYPE_UBERX = 1;
  TRIP_TYPE_POOL = 2;
}
```

This is due to Protobuf enums using C++ scoping rules. This results in it not being
possible to have two enums with the same value. For example, the following is not valid
Protobuf, regardless of file structure.

```proto
syntax = "proto3";

package uber.trip.v1;

enum Foo {
  CAR = 0;
}

enum Bar {
  // Invalid! There is already a CAR enum value in uber.trip.v1.Foo.
  CAR = 0;
}
```

Compiling this file will result in the following errors from `protoc`.

```
uber/trip/v1/trip.proto:10:3:"CAR" is already defined in "uber.trip.v1".
uber/trip/v1/trip.proto:10:3:Note that enum values use C++ scoping rules, meaning that enum values
are siblings of their type, not children of it.  Therefore, "CAR" must be unique within
"uber.trip.v1", not just within "Bar".
```

  2. All enum values must have a 0 `INVALID` value.

For example, for an enum `TripType`, there must be `TRIP_TYPE_INVALID = 0;`.


  3. The invalid value carries no semantic meaning, and if a value can be purposefully
  unset, i.e. you think a value should be purposefully null over the wire, then
  there should be a `UNSET` value as the 1 value.

For example, for an enum `TripType`, you may add a value `TRIP_TYPE_UNSET = 1;`.

Protobuf (proto3 to be specific) does not expose the concept of set vs. unset integral fields (of
which enums are), as a result it is possible to create a empty version of a message and
accidentally create the impression that an enum value was set by the caller. This can lead to hard
to find bugs where the default zero value is being set without the caller knowingly doing so. You
may be thinking - but it is super useful to just be able to assume my default enum option, just
like I want an integer field called count to default to 0 without setting it explicitly. However,
Enum values are not integers, they are just represented as them in the Protobuf description. Take
for example the following enum:

```proto
// This is not a valid example.
enum Shape {
    SHAPE_CIRCLE = 0;
    SHAPE_RECTANGLE = 1;
}
```

In this case a consumer of this Protobuf message might forget to set any `Shape` fields that exist,
and as a result the default value of `SHAPE_CIRCLE` will be assumed. This is dangerous and creates
hard to track down bugs.

Following similar logic to our INVALID case, we don't want information in messages to be implied,
we want signal to be stated with intention. If you have a case where you want `UNSET` to be a
semantic concept, then this value must be explicitly set. For example:

```proto
// The traffic light color.
enum TrafficLightColor {
    TRAFFIC_LIGHT_COLOR_INVALID = 0;
    TRAFFIC_LIGHT_COLOR_UNSET = 1;
    TRAFFIC_LIGHT_COLOR_GREEN = 2;
    TRAFFIC_LIGHT_COLOR_YELLOW = 3;
    TRAFFIC_LIGHT_COLOR_RED = 4;
}
```

It's tempting to use `UNSET` as a default value, but then again we risk the case of a user
forgetting to set the value and it being interpreted as the intentional value `UNSET`. For
consistency across our enums, if `UNSET` is a semantic value of your enum, it should have the
value 1.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Nested Enums

Nested enums are allowed, but **strongly discouraged.**

While allowed, **a good general policy is to always use unnested enums.**

Nested enums should not be referenced outside of their encapsulating message.

The following is valid but discouraged.

```proto
// A traffic light.
//
// Discouraged.
message TrafficLight {
  // A traffic light color.
  enum Color {
    COLOR_INVALID = 0;
    COLOR_UNSET = 1;
    COLOR_GREEN = 2;
    COLOR_YELLOW = 3;
    COLOR_RED = 4;
  }
  string id = 1;
  Color current_color = 2;
}
```

Note that the enum value prefix follows the same convention whether nested or unnested.

Of special note: For the V1 Style Guide, enums had their nesting types prefixed as well,
so in the above example you would have had `TRAFFIC_LIGHT_COLOR_INVALID`. This was dropped
for the V2 Style Guide.

While the above example is valid, you should not reference a `TrafficLight.Color` outside of the
`TrafficLight` message as a matter of convention - if referencd outside of `TrafficLight`, the enum
has meaning in other contexts, which means it should be top-level. If you need to reference an enum
outside of a message, instead do the following.

```proto
// A traffic light color.
enum TrafficLightColor {
    TRAFFIC_LIGHT_COLOR_INVALID = 0;
    TRAFFIC_LIGHT_COLOR_UNSET = 1;
    TRAFFIC_LIGHT_COLOR_GREEN = 2;
    TRAFFIC_LIGHT_COLOR_YELLOW = 3;
    TRAFFIC_LIGHT_COLOR_RED = 4;
}

// A traffic light.
message TrafficLight {
  string id = 1;
  TrafficLightColor current_color = 2;
}
```

Only use nested enums when you are sure that for the lifetime of your API, the enum value will not
be used outside of the message. In the above example, there could easily be situations where we
want to reference `TrafficLightColor` in other messages in the future.

```proto
// Statistics on a traffic light color.
message TrafficLightColorStats {
  string traffic_light_id = 1;
  TrafficLightColor traffic_light_color = 2;
  google.protobuf.Timestamp last_active_time = 3;
  google.protobuf.Duration total_duration = 4;
}
```

In most cases, you cannot be sure that you will never want to use an enum in another message,
and there is no cost to having an enum be unnested.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Messages

Messages should always be `PascalCase`.

All messages require a comment that contains at least one complete sentence, with the exception of
request and response messages as described in the [Services](#services) section. See the
[Documentation](#documentation) section for more details.

Messages should be used to represent logical entities that have semantic meaning. They should not
be used to group common fields that carry no meaning themselves.

The following is an example of what not to do.

```proto
// Do not do this.
message CommonFields {
  string id = 1;
  google.protobuf.Timestamp updated_time = 2;
}

// Do not do this.
message Vehicle {
  CommonFields common_fields = 1;
  string vin = 2;
}

// Do not do this.
message User {
  CommonFields common_fields = 1;
  string name = 2;
}
```

This is commonly done to simplify server-side implementations of Protobuf APIs. There are sometimes
operations you want to do on request types that require a common set of fields, so it seems to make
sense to group them as messages in your Protobuf schema.

You should optimize your Protobuf schema for the simplicity and semantic value of the API, not for
for your server-side implementations. `CommonFields` has no semantic meaning - it's a
pair of a couple fields that are part of messages that do have semantic meaning. A user of your
API does not care that these fields are common, they only care what a `Vehicle` is, and what
a `User` is. Thinking of this as JSON, the following has no meaning.

```json
{
  "common_fields": {
    "id": "asdvasd",
    "updated_time": "timestamp",
  },
  "vin": "asdvasdv"
}
```

The key `common_fields` has no meaning, `id` and `updated_time` are fields of a vehicle. Therefore,
the following is what you would want as a consumer of the API:

```json
{
  "id": "asdvasd",
  "updated_time": "timestamp",
  "vin": "asdvasdv"
}
```

In our example, you would copy the fields between `Vehicle` and `User`.

```proto
// A vehicle.
message Vehicle {
  string id = 1;
  google.protobuf.Timestamp updated_time = 2;
  string vin = 3;
}

// A user.
message User {
  string id = 1;
  google.protobuf.Timestamp updated_time = 2;
  string name = 3;
}
```

Messages that will always contain a single field are strongly discouraged. Sometimes, it is
tempting to use messages to confer type information.

```proto
// Do not do this.
message VehicleID {
  string value = 1;
}
```

The above is generally done because Protobuf happens to generate specific struct or class types in
each language, and therefore have semantic meaning. However, this is not how Protobuf values are
intended to be typed, and this can explode in a myriad of ways. To be consistent in our User
example, we would need to do the following.

```proto
// Do not do this.
message UserID {
  string value = 1;
}

// Do not do this.
message Name {
  string value = 1;
}

// Do not do this.
message User {
  UserID id = 1;
  Name first_name = 2;
  Name last_name = 3;
  repeated Name middle_names = 4;
}
```

When taken further, this becomes nightmarish to work with both in terms of internal
implementations, and from an API standpoint. The code in most languages will require
so many Builders or struct initializations just to set single values that it becomes
extremely tedious to iterate on an API. While less important, the JSON structure
also explodes.

```json
{
  "id": {
    "value": "sdfvsavassava"
  },
  "first_name": {
    "value": "John"
  },
  "last_name": {
    "value": "Smith"
  },
  "middle_names": [
    {
      "value": "Foo"
    },
    {
      "value": "Bar"
    }
  ]
}
```

As compared to:

```json
{
  "id": "sdfvsavassava",
  "first_name": "John",
  "last_name": "Smith",
  "middle_names": [
    "Foo",
    "Bar"
  ]
}
```

If a message will always be a single value, prefer that single value for fields. You will thank us
later.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Message Fields

Message field names should always be `lower_snake_case`.

While there is no enforcement of this, we recommend that message field names match their type in
lieu of a name that makes more sense. The following is an example of this pattern.

```proto
// A vehicle
message VehicleState {
  string vehicle_id = 1;
  // The field name matches the type PositionEstimate.
  PositionEstimate position_estimate = 2;
  // The field name is the plural of the type VehicleFeature.
  repeated VehicleFeature vehicle_features = 3;
  // The map is from key type to value type.
  map<string, Order> order_id_to_order = 4;
  // Timestamps are the exception, see below.
  google.protobuf.Timestamp time = 5;
}
```

The built-in message field option `json_name` should never be used.

The following additional naming rules apply.

- Field names cannot contain `descriptor`. This causes a collision in Java-generated code.
- Field names cannot contain `file_name`, instead using `filename`. This is just for consistency.
- Field names cannot contain `file_path`, instead using `filepath`. This is just for consistency.
- Fields of type `google.protobuf.Timestamp` should be named `time` or end in `_time`. For example,
  `foo_time`.
- Fields of type `google.protobuf.Duration` should be named `duration` or end in `_duration`. For
  example, `foo_duration`.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Oneofs

Oneof names should always be `lower_snake_case`.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Nested Messages

Nested messages are allowed, but **strongly discouraged.**

While allowed, **a good general policy is to always use unnested messages.**

Nested messages should not be referenced outside of their encapsulating message.

The following is valid but discouraged.

```proto
// A Vehicle.
//
// Discouraged.
message Vehicle {
  // A vehicle type.
  message Type {
    // Should probably be an enum.
    string make = 1;
    string model = 2;
    uint32 year = 3;
  }
  string id = 1;
  Type type = 2;
}
```

While the above example is valid, you should not reference a `Vehicle.Type` outside of the
`Vehicle` message as a matter of convention - if referencd outside of `Vehicle`, the message has
meaning in other contexts, which means it should be top-level. If you need to reference a message
outside of a message, instead do the following.

```proto
// A vehicle type.
message VehicleType {
  // Should probably be an enum.
  string make = 1;
  string model = 2;
  uint32 year = 3;
}

// A vehicle.
message Vehicle {
  string id = 1;
  VehicleType vehicle_type = 2;
}
```

Only use nested messages when you are sure that for the lifetime of your API, the message value
will not be used outside of the message. In the above example, there could easily be situations
where we want to reference `VehicleType` in other messages in the future.

```proto
// Statistics on a vehicle type.
message VehicleTypeStats {
  VehicleType vehicle_type = 1;
  uint64 number_made = 2;
}
```

In most cases, you cannot be sure that you will never want to use a message in another message,
and there is no cost to having a message be unnested.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Services/RPCs

Services should always be `PascalCase`. RPCs should always be `PascalCase.`

Services should always be suffixed with `API`. For example, `TripAPI` or `UserAPI`. This is for
consistency.

All services and RPCs require a comment that contains at least one complete sentence. See the
[Documentation](#documentation) section for more details.

Please read the above [File Structure](#file-structure) section carefully. Each service should
be in it's own file named after the service. For example, `TripAPI` should be in a file named
`trip_api.proto`.

Every RPC request and response should be unique to the RPC, and named after the RPC. As per the
[File Structure](#file-structure) section, the request and response messages should be in the
same file as the service, underneath the service definition and in order of the RPCs defined
in the service.

The unique request/response requirement is for compatibility as you iterate your RPC definitions.
If one were to use the same request or response message for different RPCs, and then need to add or
want to deprecate a field, each RPC that uses this request and response message would be affected,
regardless of whether or not the new field applied to each RPC, or whether the deprecated field
was deprecated for each RPC. While this pattern can cause some duplication, it's become the "gold
standard" for Protobuf APIs.

The following is an example of such a pattern.

```proto
// Handles interaction with trips.
service TripAPI {
  // Get the trip specified by the ID.
  rpc GetTrip(GetTripRequest) returns (GetTripResponse);
  // List the trips for the given user before a given time.
  //
  // If the start index is beyond the end of the available number
  // of trips, an empty list of trips will be returned.
  // If the start index plus the size is beyond the available number
  // of trips, only the number of available trips will be returned.
  rpc ListUserTrips(ListUserTripsRequest) returns (ListUserTripsResponse);
}

message GetTripRequest {
  string id = 1;
}

message GetTripResponse {
  Trip trip = 1;
}

message ListUserTripsRequest {
  string user_id = 1;
  google.protobuf.Timestamp before_time = 2;
  // The start index for pagination.
  uint64 start = 3;
  // The maximum number of trips to return.
  uint64 max_size = 4;
}

message ListUserTripsResponse {
  repeated Trip trips = 1;
  // True if more trips are available.
  bool next = 2;
}
```

Note that request and response types do not need documentation comments, as opposed to all other
messages.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### Streaming RPCs

Streaming RPCs are allowed, but **strongly discouraged**.

If you'd like to enforce this, add the following to your `prototool.yaml`.

```yaml
lint:
  group: uber2
  rules:
    add:
      - RPCS_NO_STREAMING
```

Streaming RPCs seem enticing at first, but they push RPC framework concerns to the application
level, which results in inconsistent implementations and behavior, which results in lower
reliability. It is rare that a problem solved by streaming cannot be solved in a unary manner.

Spencer Nelson wrote a great summary of the issues with streaming RPCs in
[this GitHub Issue comment](https://github.com/twitchtv/twirp/issues/70#issuecomment-470367807),
which we paraphrase here:

Unary calls, while limiting, prevent many problems being created if you only give unary RPC
capabilities to an inexperienced developer as part of your tool kit. While you can make mistakes in
API design and message design, those mistakes stay local, they largely aren't architectural or
infrastructural.

Streams are different. They imply expectations about how connection state is managed. Load
balancing them is really hard, which means that a decision to use streams for an API has
ramifications that ripple out to the infrastructure of a system.

How plausible is it that users trip on that pothole? Unfortunately, it is quite likely. Streaming
is the sort of feature whose downsides can be hard to see at first. "Lower latency and the ability
to push data to the client? Sign me up!" But people could easily reach for it too early in order
to future-proof their APIs, "just in case we need streams later," and walk themselves into an
architectural hole which is very difficult to get out of.

Most simple streaming RPCs can be implemented in terms of request-response RPCs, so long as they
don't have extreme demands on latency or number of open connections (although HTTP/2 request
multiplexing mostly resolves those issues anyway). Pagination and polling are the obvious tools
here, but even algorithms like rsync are surprisingly straightforward to implement in a
request/response fashion, and probably more efficient than you think if you're using HTTP/2,
since the transport is streaming anyway.

Sometimes you really do need streams for your system, like if you are replicating a data stream.
Nothing else is really going to work, there. But Protobuf APIs do not need to be all things for all
use-cases. There is room for specialized solutions, but streams are a special thing, and they have
resounding architectural impact.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

### HTTP Annotations

[HTTP annotations](https://github.com/googleapis/googleapis/blob/master/google/api/annotations.proto)
for use with libraries such as [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway)
are allowed, but **strongly discouraged**.

If you'd like to enforce this, add the following to your `prototool.yaml`.

```yaml
lint:
  group: uber2
  rules:
    add:
      - RPC_OPTIONS_NO_GOOGLE_API_HTTP
```

HTTP annotations provide an alternative call path for RPCs that is not structured per the main
Protobuf schema. One of the benefits of Protobuf is having this structure, and generated clients
which properly call each endpoint.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Naming

Some names are disallowed from all Protobuf types. The following names are not allowed to be part
of any package, option, message, message field, enum, enum value, service, or RPC, with any
capitalization.

- `common` - Common has no semantic meaning. See the above discussions on not having common
  messages. Use a name that reflects the actual meaning of what you are representing instead.
- `data` - The name "data" is a superfluous decorator. All Protobuf types are data. Use of
  "data" also causes singular vs plural issues during iteration, as the singluar of "data" is
  "datum". If you must have a type that needs such a decorator, use "info" instead.
- `uuid` - We use "id" instead of "uuid" for purposes of consistency. An ID is meant to be
  a UUID unless otherwise specified.

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**

## Documentation

All comments should use `//` and not `/* */`. This is for consistency.

Comments should never be inline, instead always being above the type.

```proto
// A foo. Note that "A Foo." is a complete sentence as we define it.
message Vehicle {
  // Comment here.
  string id = 1; // Do not do inline comments.
}
```

While not enforced, aim for a 120 character maximum length for comments.

You must document your messages, enums, services, and RPCs. These types require at least
one complete sentence, represented as starting with a capital letter and ending with a period.

The following are some general recommendations.

**Message Documentation**

Description of message. Consider including:
- Assumptions and requirements (e.g. "This polygon must have its points in a counter-clockwise order.").

**Message Field Documentation**

Description of field. Not required, but recommended. Consider including:
- Assumptions and requirements.
- What happens when the field is left blank. Does it default to a specific value or throw an invalid argument error?

**Service Documentation**

Explanation of what the service is intended to do/not do. Consider including:

- Advantage or use cases.
- Related services (mark with the `@see` annotation).

**RPC Documentation**

Explanation of what the RPC is intended to do/not do. Consider including:

- Advantages or use cases (e.g."Useful when you want to send large volumes and don't care about latency.").
- Side effects (e.g."If a feature with this ID already exists, this method will overwrite it.").
- Performance considerations (e.g."Sending your data in chunks of X size is more efficient.").
- Pre-requisites (e.g. "You must complete registration using the X method before calling this one.").
- Post-requisites (e.g. "Clean up the registered resource after use to free up resources.").

**[⬆ Back to top](#uber-protobuf-style-guide-v2)**
