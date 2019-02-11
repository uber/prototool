# Prototool Docker Image

This directory provides a Docker image with `prototool, `protoc`, and common Protobuf plugins pre-installed. As of
this writing, the resulting image is around 141MB. This provides a consistent environment to generate your Protobuf
stubs.

This is in early development.

## Included

| Name | Version | Binaries | Comments |
| --- | --- | --- | --- |
| [protoc] | 3.6.1 | `protoc` | Well-Known Types copied to `/usr/local/incluide` |

[protoc]: https://github.com/protocolbuffers/protobuf
