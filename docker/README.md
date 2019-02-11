# Prototool Docker Image

This directory provides a Docker image with `prototool`, `protoc`, and common Protobuf plugins pre-installed. As of
this writing, the resulting image is around 141MB. This provides a consistent environment to generate your Protobuf
stubs.

This is in early development.

## Included

| Name | Version | Binaries |
| --- | --- | --- |
| prototool | (varies) | prototool |
| [protoc] | 3.6.1 | protoc |
| [grpc] | 1.18.0 | grpc_cpp_plugin<br>grpc_csharp_plugin<br>grpc_node_plugin<br>grpc_objective_c_plugin<br>grpc_php_plugin<br>grpc_python_plugin<br>grpc_ruby_plugin |
| [golang/protobuf] | 1.2.0 | protoc-gen-go |
| [gogo/protobuf] | 1.2.0 | protoc-gen-gofast<br>protoc-gen-gogo<br>protoc-gen-gogofast<br>protoc-gen-gogofaster<br>protoc-gen-gogoslick |
| [grpc-gateway] | 1.7.0 | protoc-gen-grpc-gateway<br>protoc-gen-swagger |
| [grpc-web] | 1.0.3 | protoc-gen-grpc-web |
| [twirp] | 5.5.1 | protoc-gen-twirp<br>protoc-gen-twirp_python |
| [yarpc] | 1.36.1 | protoc-gen-yarpc-go |

The Well-Known Types are copied to `/usr/include`.


[protoc]: https://github.com/protocolbuffers/protobuf
[grpc]: https://github.com/grpc/grpc
[golang/protobuf]: https://github.com/golang/protobuf
[gogo/protobuf]: https://github.com/gogo/protobuf
[grpc-gateway]: https://github.com/grpc-ecosystem/grpc-gateway
[grpc-web]: https://github.com/grpc/grpc-web
[twirp]: https://github.com/twitchtv/twirp
[yarpc]: https://github.com/yarpc/yarpc-go
