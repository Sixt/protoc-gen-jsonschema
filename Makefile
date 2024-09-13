PROTO_PATH ?= "internal/converter/testdata/proto"

default: build

.PHONY: build
build:
	mkdir -p bin
	go build -o bin/protoc-gen-jsonschema cmd/protoc-gen-jsonschema/main.go

.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w -local github.com/sixt/protoc-gen-jsonschema .

.PHONY: install
install:
	go install github.com/sixt/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema

.PHONY: build_linux
build_linux:
	GOOS=linux GOARCH=amd64 go build -o protoc-gen-jsonschema.linux-amd64

.PHONY: samples
samples: build
	mkdir -p jsonschemas
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfMessages.proto || echo "No messages found (ArrayOfMessages.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfObjects.proto || echo "No messages found (ArrayOfObjects.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfPrimitives.proto || echo "No messages found (ArrayOfPrimitives.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas -I. --proto_path=${PROTO_PATH} ${PROTO_PATH}/Enumception.proto || echo "No messages found (Enumception.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_additional_properties:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/NestedMessage.proto || echo "No messages found (NestedMessage.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/NestedObject.proto || echo "No messages found (NestedObject.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/PayloadMessage.proto || echo "No messages found (PayloadMessage.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/SeveralEnums.proto || echo "No messages found (SeveralEnums.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/SeveralMessages.proto || echo "No messages found (SeveralMessages.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Timestamp.proto || echo "No messages found (Timestamp.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=all_fields_required:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/PayloadMessage2.proto || echo "No messages found (PayloadMessage2.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=json_fieldnames:jsonschemas -I. --proto_path=${PROTO_PATH} ${PROTO_PATH}/JSONFields.proto || echo "No messages found (JSONFields.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/ArrayOfEnums.proto || echo "No messages found (SeveralMessages.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Maps.proto || echo "No messages found (Maps.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/MessageWithComments.proto || echo "No messages found (MessageWithComments.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Proto2Required.proto || echo "No messages found (Proto2Required.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Proto2NestedMessage.proto || echo "No messages found (Proto2NestedMessage.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/GoogleValue.proto || echo "No messages found (GoogleValue.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/GoogleInt64Value.proto || echo "No messages found (GoogleInt64Value.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/GoogleInt64ValueDisallowString.proto || echo "No messages found (GoogleInt64ValueDisallowString.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/GoogleInt64ValueAllowNull.proto || echo "No messages found (GoogleInt64ValueAllowNull.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=disallow_bigints_as_strings,allow_null_values:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/GoogleInt64ValueDisallowStringAllowNull.proto || echo "No messages found (GoogleInt64ValueDisallowStringAllowNull.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=enforce_oneof:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/OneOf.proto || echo "No messages found (OneOf.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=all_fields_required:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/Proto2NestedObject.proto || echo "No messages found (Proto2NestedObject.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/WellKnown.proto || echo "No messages found (WellKnown.proto)"
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/NoPackage.proto
	protoc --plugin=bin/protoc-gen-jsonschema --jsonschema_out=messages=[MessageKind10+MessageKind11+MessageKind12]:jsonschemas --proto_path=${PROTO_PATH} ${PROTO_PATH}/TwelveMessages.proto || echo "No messages found (TwelveMessages.proto)"

.PHONY: test
test:
	go test ./... -cover -v
