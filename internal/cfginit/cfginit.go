// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package cfginit contains the template for prototool.yaml files, as well
// as a function to generate a prototool.yaml file given a specific protoc
// version, with or without commenting out the remainder of the options.
package cfginit

import (
	"bytes"
	"html/template"
)

var tmpl = template.Must(template.New("tmpl").Parse(`# Paths to exclude when searching for Protobuf files.
{{.V}}excludes:
{{.V}}  - path/to/a
{{.V}}  - path/to/b/file.proto

# Protoc directives.
protoc:
  # The Protobuf version to use from https://github.com/protocolbuffers/protobuf/releases.
  # By default use {{.ProtocVersion}}.
  # You probably want to set this to make your builds completely reproducible.
  version: {{.ProtocVersion}}

  # Additional paths to include with -I to protoc.
  # By default, the directory of the config file is included,
  # or the current directory if there is no config file.
  {{.V}}includes:
  {{.V}}  - ../../vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis


  # If not set, compile will fail if there are unused imports.
  # Setting this will ignore unused imports.
  {{.V}}allow_unused_imports: true

# Create directives.
{{.V}}create:
  # List of mappings from relative directory to base package.
  # This affects how packages are generated with create.
  {{.V}}packages:
    # This means that a file created "foo.proto" in the current directory will have package "bar".
    # A file created "a/b/foo.proto" will have package "bar.a.b".
    {{.V}}- directory: .
    {{.V}}  name: bar
    # This means that a file created "idl/code.uber/a/b/c.proto" will have package "uber.a.b".
    {{.V}}- directory: idl/code.uber
    {{.V}}  name: uber

# Lint directives.
{{.V}}lint:
  # Linter files to ignore.
{{.V}}  ignores:
{{.V}}    - id: RPC_NAMES_CAMEL_CASE
{{.V}}      files:
{{.V}}        - path/to/foo.proto
{{.V}}        - path/to/bar.proto
{{.V}}    - id: SYNTAX_PROTO3
{{.V}}      files:
{{.V}}        - path/to/foo.proto

  # Linter rules.
  # Run prototool list-all-linters to see all available linters.
{{.V}}  rules:
    # Determines whether or not to include the default set of linters.
{{.V}}    no_default: true

    # The specific linters to add.
{{.V}}    add:
{{.V}}      - ENUM_NAMES_CAMEL_CASE
{{.V}}      - ENUM_NAMES_CAPITALIZED

    # The specific linters to remove.
{{.V}}    remove:
{{.V}}      - ENUM_NAMES_CAMEL_CASE

# Code generation directives.
{{.V}}generate:
  # Options that will apply to all plugins of type go and gogo.
{{.V}}  go_options:
    # The base import path. This should be the go path of the prototool.yaml file.
    # This is required if you have any go plugins.
{{.V}}    import_path: uber/foo/bar.git/idl/uber

    # Extra modifiers to include with Mfile=package.
{{.V}}    extra_modifiers:
{{.V}}      google/api/annotations.proto: google.golang.org/genproto/googleapis/api/annotations
{{.V}}      google/api/http.proto: google.golang.org/genproto/googleapis/api/annotations

  # The list of plugins.
{{.V}}  plugins:
      # The plugin name. This will go to protoc with --name_out, so it either needs
      # to be a built-in name (like java), or a plugin name with a binary
      # protoc-gen-name.
{{.V}}    - name: gogo

      # The type, if any. Valid types are go, gogo.
      # Use go if your plugin is a standard Golang plugin
      # that uses github.com/golang/protobuf imports, use gogo
      # if it uses github.com/gogo/protobuf imports. For protoc-gen-go
      # use go, For protoc-gen-gogo, protoc-gen-gogoslick, etc, use gogo.
{{.V}}      type: gogo

      # Extra flags to specify.
      # The only flag you will generally set is plugins=grpc for Golang.
      # The Mfile=package flags are automatically set.
      # ** Otherwise, generally do not set this unless you know what you are doing. **
{{.V}}      flags: plugins=grpc

      # The path to output generated files to.
      # If the directory does not exist, it will be created when running generation.
      # This needs to be a relative path.
{{.V}}      output: ../../.gen/proto/go

      # Optional override for the plugin path. For example, if you set set path to
      # /usr/local/bin/gogo_plugin", prototool will add the
      # "--plugin=protoc-gen-gogo=/usr/local/bin/gogo_plugin" flag to protoc calls.
{{.V}}      path: /usr/local/bin/gogo

{{.V}}    - name: yarpc-go
{{.V}}      type: gogo
{{.V}}      output: ../../.gen/proto/go

{{.V}}    - name: grpc-gateway
{{.V}}      type: go
{{.V}}      output: ../../.gen/proto/go

{{.V}}    - name: java
{{.V}}      output: ../../.gen/proto/java`))

type tmplData struct {
	V             string
	ProtocVersion string
}

// Generate generates the data.
//
// Set uncomment to true to uncomment the example settings.
func Generate(protocVersion string, uncomment bool) ([]byte, error) {
	tmplData := &tmplData{
		ProtocVersion: protocVersion,
	}
	if !uncomment {
		tmplData.V = "#"
	}
	buffer := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buffer, tmplData); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
