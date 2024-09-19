.PHONY: build install, build_linux samples test

build:
	mkdir -p bin
	go build -o bin/protoc-gen-jsonschema cmd/protoc-gen-jsonschema/main.go

install:
	GO111MODULE=on go get -u github.com/sixt/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema && go install github.com/sixt/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema

build_linux:
	GOOS=linux GOARCH=amd64 go build -o protoc-gen-jsonschema.linux-amd64

PROTO_PATH ?= "internal/converter/testdata/proto"
samples: build
	mkdir -p jsonschemas
	PATH=./bin:$$PATH; protoc --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfMessages.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfObjects.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfPrimitives.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_additional_properties:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Enumception.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_additional_properties:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ImportedEnum.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_additional_properties:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/NestedMessage.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/NestedObject.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/PayloadMessage.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/SeveralEnums.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/SeveralMessages.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfEnums.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Maps.proto
	PATH=./bin:$$PATH; protoc --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/MessageWithComments.proto
	PATH=./bin:$$PATH; protoc -I /usr/include --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/WellKnown.proto

test:
	go test ./... -cover
