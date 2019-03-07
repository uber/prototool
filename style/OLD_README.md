# Uber Protobuf Style Guide

## Table of Contents
(insert here)

## Source File Setup
Uber splits protobuf constructs into two types of files: **Service files** and **Supporting files**.
### Service Files
Contain services, RPCs, and their associated request and response messages.
- **Naming**: Name of the service followed by `_api` (e.g. `MapAPIService_api.proto`) .
- **Organization**: Service Files contain the following elements in this order:
   1. License Header (if applicable)
   2. Syntax
   3. Package
   4. Package Options (alphabetized)
   5. Imports (alphabetized)
   6. Services & RPCs
   7. Request and response message pairs (ordered by their appearance in the rpcs)


### Supporting Files
Contain messages and enums that are used by one or more services.
- **Naming**: Should *not* have `_api` at the end (e.g. `file_name_here.proto`).
- **Organization**: Supporting Files contain the following elements in this order:
  1. License Header (if applicable)
  2. Syntax
  3. Package
  4. Package Options (alphabetized)
  5. Imports (alphabetized)
  6. Messages
  7. Enums

Place one blank line between each element (license header, syntax, package, package options, and imports).
```protobuf
//syntax is always "proto3".
syntax = "proto3";

package style.uber;

// TODO this is out of date
// Always specify the java and golang options listed below. The java options match those specified in
// Google Cloud's documentation: https://cloud.google.com/apis/design/file_structure
// The go package is always $(basename PACKAGE)pb.
// Do not use the "long-form" package name with a directory path.
option go_package = "uberpb";
// java_multiple_files is always true.
option java_multiple_files = true;
// java_outer_classname is the CamelCase file name without the extension, followed by Proto.
option java_outer_classname = "UberProto";
// The java package is always com.PACKAGE.
option java_package = "com.style.uber";

//Imports should have no blank lines between them.
import "dep/dep.proto";
// Import Google's well-known types directly from "google/protobuf".
import "google/protobuf/empty.proto";
```

## Services
- Use CamelCase for Service names and end them with "Service" (e.g. `HelloService`).
- Use ; to terminate the method definition when there are no RPC options and {} when there are RPC options.


```protobuf
// Without RPC options
service HelloService;

// With RPC options
service GoodbyeService {
  rpc SayGoodbye (SayGoodbyeRequest) returns (SayGoodbyeResponse);
}
```

## RPCs
- RPC names should be in CamelCase and begin with a verb (e.g. `StreamGalaxies`).
- Each RPC must have its own request and response message, even if the messages are empty. The request and response names should always match the RPC name.
  > **Why?:** This is for backwards compatibility.

```protobuf
// Example Service and RPCs.
// The RPC has it's own request and response messages.

service PlanetService {
  rpc GetPlanet(GetPlanetRequest) returns (GetPlanetResponse);
}

message GetPlanetRequest {
  string planet_id = 1;
}

message GetPlanetResponse {
  Planet planet = 1;
}
```

## Messages
- Use CamelCase for message names (e.g. `MapFeature`).

- Request and response names should always match the rpc name.
```protobuf
rpcStreamQuasars(stream StreamQuasarsRequest) returns (StreamQuasarsResponse);
message StreamGalaxiesRequest{}
```
- Request and response messages should not have any nested messages or enums.

- If a message is empty, do not include a line break after the opening bracket.
```protobuf
// No line break after opening bracket because message is empty.
message StreamGalaxiesRequest{}
```

### Message Fields
- Use all lowercase for message fields names, with underscores between each word.
```protobuf
sint32 longitude micros = 3;
```

- Use plural case for repeated message fields.
> **Why?:** Traditionally for Protobuf, repeated fields use singular case. We prefer plural case because we have found that singular is more confusing and that few developers actually used singular case in practice anyway.

  ```protobuf
repeated string planet_ids = 3;
```

- Mark deprecated fields with `[deprecated = true]`. Include one space on each side of the equals sign.
> **Why?:** We mark fields as deprecated, rather than removing them and setting them as reserved, because we want to disallow reusing field names for JSON compatibility. By keeping the field and marking it as deprecated, we make it impossible to reuse either the field tag or the field name.

  ```protobuf
string foo = 1 [deprecated = true];
```
- Use the "right" primitive type for the situation, regardless of generated code in the particular target language. For example, use uint32 for ports and not int32, uint64, int64, etc. See: https://developers.google.com/protocol-buffers/docs/proto3#scalar
```protobuf
// Here an sint32 is used instead of int32 because there is a high
// probability of having a negative value and 32 bits instead of 64,
// as by definition, this will never exceed 32 bits.
sint32 latitude_micros = 3;
sint32 longitude_micros = 4;
```

- Use the string type for ID fields.
> **Why?:** Repos have sometimes had wrapper message types for IDs, but this provides little value in practice and causes a lot of code uncleanliness.

- If there is a type enum associated with a message, the name of the field should be "type."

- The ID field should use the first tag, unless there is a type field. When there is a type field, type should use the first tag and ID should use the second tag.
```protobuf
// Message with just an ID.
// ID is always a string.
message Foo {
  string foo_id = 1;
}
// Message with a type and an ID.
// Type is now the first tag.
message Dog {
  DogType type = 1;
  string dog_id = 2;
}
```

### Nested Messages
You can nest messages and enums, *except* in request and response messages. Nesting messages and enums is appropriate in cases where the inner message has no meaning or purpose outside of the outer message.
>**Warning:** This affects the names of generated types and may add a great amount of verbosity, so do this at your own discretion.


```protobuf
// Bar has an embedded Type enum and an ID.
message Bar {
    enum Type {
      BAR_TYPE_INVALID = 0;
      BAR_TYPE_UNSET = 1;
      BAR_TYPE_REMOTE_CONTROL = 2;
      BAR_TYPE_FAN = 3;
    }
Type type = 1;
string id = 2;
}
```



## Enums
- Use CamelCase for enum names (e.g. `TrafficLight`).

### Enum Values
- Use all caps for enum value names with underscores between each word. Also, include the enum name as a prefix (e.g. `TRAFFIC_LIGHT_RED`)
> **Why?** Using the enum name as a prefix is necessary for C++ scoping rules.

- Always include a zero value with suffix `_INVALID` (e.g. `TRAFFIC_LIGHT_INVALID = 0;`). However, if you need to denote an actual null value over the wire, set this because `_INVALID` is not a valid value to check against.
> **Why?**
> Protocol buffers v3+  does not expose the concept of set vs. unset integral
> fields (which enums are). As a result, it is possible to create an empty version
> of a message and accidentally create the impression that the caller set an enum value.
> This can lead to hard to find bugs where the 0 enum option
> is being set without the caller's knowledge. You may be thinking, "But it is super useful to
> just be able to assume my default enum option. For example, if I had a field
> called count, I'd want it to default to 0 without setting it explicitly." The thing is,
> enums are not integers, they are just represented as integers the proto
> description. For example:
```protobuf
// Don't do this.
// A consumer of this message might forget to set any Shape
// fields that exist, and then the default value of Circle will
// be assumed. This is dangerous and creates hard to track down bugs.
enum Shape {
SHAPE_CIRCLE = 0;
SHAPE_RECTANGLE = 1;
}
```

- If you want to denote a purposefully unset value, include one value with the suffix `_UNSET`. If you include an `_UNSET` enum value, it must be numbered 1 (e.g. ``{ENUM_TYPE}_UNSET = 1);``.
> **Why?** If you make `_UNSET` the default, it may appear as if the user has intentionally set the value to `_UNSET` when really, they just forgot to set a value for the enum.
```protobuf
// Correct use of _INVALID and _UNSET
enum TrafficLight {
  TRAFFIC_LIGHT_INVALID = 0;
  TRAFFIC_LIGHT_UNSET = 1;
  TRAFFIC_LIGHT_GREEN = 2;
  TRAFFIC_LIGHT_YELLOW = 3;
  TRAFFIC_LIGHT_RED = 4;
}
```



## Code Comments
The Uber linter enforces comments at the top of each:
- File (see [File Overview Comments](#file-overview-comments-required))
- Service (see [Service Description](#service-description-required))
- RPC (see [RPC Description](#rpc-description-required))
- Message (see [Message Description](#message-description-required))
- Enum (see [Enum Description](#enum-description-required))

In addition, we recommend placing comments on top of message fields (see [Message Field Description](#message-field-description-optional)) and enum values (see [Enum Field Description](#enum-field-description-optional)) as is appropriate.

Reviewers can also block your code because of comments that are low-quality or formatted incorrectly. Below, we provide guidance about how to write useful comments in Uber style.


### Comment Format
- **Format**: Comments begin with // and one space. Do not use /**/.
- **Placement**: Comments go above the construct they describe with no blank lines in between.
- **Line Length**: Aim for 120 characters maximum.
- **Tabs**: Use two spaces for tab.
- **Phrasing**: Begin comments with a verb phrase in the third person, where possible.
  - Messages: Use "Represents".
  - Enums: Use "Lists".
- **Capitalization**: Start comments with a capital letter.
- **Punctuation**: End comments with a period, even if they're a sentence fragment.

```proto
//INSERT AN EXAMPLE HERE - include a bad example of comments on the same line as code.
```

**[⬆ back to top](#table-of-contents)**



### File Overview Comments (Required)
Overview and contextual information placed above the syntax line. Consider including:
- Description of what is included in the file.
- Who is intended to use the file (e.g. "third-party vendors")
- Brief explanation of jargon (e.g. "Overseer is ...")
- @see tag to indicate closely related files
- @link tag to link to external documentation (e.g. Map Feature definitions)

```protobuf
//insert example here
```
**[⬆ back to top ](#table-of-contents)**


### Service Description (Required)
Explanation of what the service is intended to do/not do. Consider including:
- Advantage or use cases
- Related services (mark with the @see annotation)

```protobuf
//insert example here
```
**[⬆ back to top ](#table-of-contents)**


### RPC Description (Required)
Explanation of what the rpc is intended to do/not do. Consider including:
- Advantages or use cases (e.g."Useful when you want to send large volumes and don't care about latency.")
- Side effects (e.g."If a feature with this ID already exists, this method will overwrite it.")
- Performance considerations (e.g."Sending your data in chunks of X size is more efficient.")
- Pre-requisites (e.g. "You must complete registration using the X method before calling this one.")
- Post-requisites (e.g. "Clean up the registered resource after use to free up resources.")

```protobuf
//Write a feature to a map. If a feature with the same ID already exists,
//this method will overwrite it.
rpc WriteFeature(WriteFeatureRequest) returns (WriteFeatureResponse);
```
**[⬆ back to top ](#table-of-contents)**

### Message Description (Required)
Description of message, beginning with "Represents." Consider including:
- Assumptions and requirements (e.g. "This polygon must have its points in a counter-clockwise order.")

### Message Field Description (Optional)
**(All Message Types)** Description of field. Consider including:
- Assumptions and requirements
- What happens when the field is left blank. Does it default to a specific value or throw an invalid argument error?

```protobuf
EXAMPLE HERE
```

**(Request Types Only)** A list of all the possible errors and the cases in which those errors are thrown.
- Errors are listed after the field description.
- Each error type is on a separate line.
- Multiple cases should be formatted as a dashed list (see below).
- Error name should be preceded with an @exception annotation.
- Error name should be followed by “if” and then a description of the case(s) in which the error is thrown.

```protobuf
// @exception INVALID_ARGUMENT if the polygon has:
// - coincident points.
// - fewer than three points.
// - self-intersections.
```
**[⬆ back to top ](#table-of-contents)**

### Enum Description (Required)
Description of enum, beginning with "Lists".
- Assumptions and requirements

### Enum Field Description (Optional)
Document enum values only when their use isn’t immediately obvious from their name.
```
// Lists the possible states of a standard U.S. traffic light.
enum TrafficLight {
  TRAFFIC_LIGHT_INVALID = 0;
  TRAFFIC_LIGHT_UNSET = 1;
  TRAFFIC_LIGHT_GREEN = 2;
  TRAFFIC_LIGHT_YELLOW = 3;
  TRAFFIC_LIGHT_RED = 4;
}
```
**[⬆ back to top ](#table-of-contents)**
