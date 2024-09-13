Protobuf to JSON-Schema compiler
================================

This takes protobuf definitions and converts them into JSONSchemas, which can be used to dynamically validate JSON messages.

Useful for people who define their data using ProtoBuf, but use JSON for the "wire" format.

"Heavily influenced" by [Google's protobuf-to-BigQuery-schema compiler](https://github.com/GoogleCloudPlatform/protoc-gen-bq-schema).


Generated Schemas
-----------------

- One JSONSchema file is generated for each root-level proto message and ENUM. These are intended to be stand alone self-contained schemas which can be used to validate a payload derived from their source proto message
- Nested message schemas become [referenced "definitions"](https://cswr.github.io/JsonSchema/spec/definitions_references/). This means that you know the name of the proto message they came from, and their schema is not duplicated (within the context of one JSONSchema file at least)


Logic
-----

- For each proto file provided
  - Generates schema for each ENUM
    - JSONSchema filename deried from ENUM name
  - Generates schema for each Message
    - Builds a list of every nested message and converts them to JSONSchema
    - Recursively converts attributes and nested messages within the root message
      - Special handling for "OneOf"
      - Special handling for arrays
      - Special handling for maps
    - Injects references to nested messages
    - JSONSchema filename derived from Message name
  - Bundles these into a protoc generator response


Installation
------------

> Note: This tool requires Go 1.11+ to be installed.

Install this plugin using Go:

```sh
go install github.com/sixt/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema@latest
```


Usage
-----

> Note: This plugin requires the [`protoc`](https://github.com/protocolbuffers/protobuf) CLI to be installed.

**protoc-gen-jsonschema** is designed to run like any other proto generator.

```sh
protoc \ # The protobuf compiler
--proto_path=testdata/proto testdata/proto/ArrayOfPrimitives.proto # proto input directories and folders
```

Sample protos (for testing)
---------------------------

* Proto with a simple (flat) structure: [samples.PayloadMessage](internal/converter/testdata/proto/PayloadMessage.proto)
* Proto containing a nested object (defined internally): [samples.NestedObject](internal/converter/testdata/proto/NestedObject.proto)
* Proto containing a nested message (defined in a different proto file): [samples.NestedMessage](internal/converter/testdata/proto/NestedMessage.proto)
* Proto containing an array of a primitive types (string, int): [samples.ArrayOfPrimitives](internal/converter/testdata/proto/ArrayOfPrimitives.proto)
* Proto containing an array of objects (internally defined): [samples.ArrayOfObjects](internal/converter/testdata/proto/ArrayOfObjects.proto)
* Proto containing an array of messages (defined in a different proto file): [samples.ArrayOfMessage](internal/converter/testdata/proto/ArrayOfMessage.proto)
* Proto containing multi-level enums (flat and nested and arrays): [samples.Enumception](internal/converter/testdata/proto/Enumception.proto)
* Proto containing a stand-alone enum: [samples.ImportedEnum](internal/converter/testdata/proto/ImportedEnum.proto)
* Proto containing 2 stand-alone enums: [samples.FirstEnum, samples.SecondEnum](internal/converter/testdata/proto/SeveralEnums.proto)
* Proto containing 2 messages: [samples.FirstMessage, samples.SecondMessage](internal/converter/testdata/proto/SeveralMessages.proto)
* Proto containing 12 messages: [samples.MessageKind1 - samples.MessageKind12](internal/converter/testdata/proto/TwelveMessages.proto)


Links
-----

* [About JSON Schema](http://json-schema.org/)
* [Popular GoLang JSON-Schema validation library](https://github.com/xeipuuv/gojsonschema)
* [Another GoLang JSON-Schema validation library](https://github.com/lestrrat/go-jsschema)
