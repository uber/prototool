# Breaking Change Detector

Protobuf is a great way to represent your APIs and generate stubs in each language you develop
with. As such, Protobuf APIs should be stable so as not to break consumers across repositories.
Even in a monorepo context, making sure that your Protobuf APIs do not introduce breaking
changes is important so that different deployed versions of your services do not have
wire incompatibilities.

Prototool exposes a breaking change detector through the `prototool break check` command.
