package converter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/sixt/protoc-gen-jsonschema/internal/converter/testdata"
)

const (
	sampleProtoDirectory = "testdata/proto"
)

type sampleProto struct {
	Flags                 ConverterFlags
	ExpectedFileNames     []string
	ExpectedJSONSchema    []string
	FilesToGenerate       []string
	ObjectsToValidateFail []string
	ObjectsToValidatePass []string
	ProtoFileName         string
	TargetedMessages      []string
}

func TestGenerateJsonSchema(t *testing.T) {

	// Configure the list of sample protos to test, and their expected JSON-Schemas:
	sampleProtos := configureSampleProtos()

	// Convert the protos, compare the results against the expected JSON-Schemas:
	for name, sampleProto := range sampleProtos {
		t.Run(name, func(t *testing.T) {
			testConvertSampleProto(t, sampleProto)
		})
	}
}

func testConvertSampleProto(t *testing.T, sampleProto sampleProto) {
	t.Helper()

	// Make a Logrus logger:
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.SetOutput(os.Stderr)

	// Use the logger to make a Converter:
	protoConverter := New(logger)
	protoConverter.Flags = sampleProto.Flags

	// Open the sample proto file:
	sampleProtoFileName := fmt.Sprintf("%v/%v", sampleProtoDirectory, sampleProto.ProtoFileName)
	fileDescriptorSet := mustReadProtoFiles(t, sampleProtoDirectory, sampleProto.ProtoFileName)

	// Prepare a request:
	codeGeneratorRequest := pluginpb.CodeGeneratorRequest{
		FileToGenerate: sampleProto.FilesToGenerate,
		ProtoFile:      fileDescriptorSet.GetFile(),
	}

	// Test the TargetedMessages feature:
	if len(sampleProto.TargetedMessages) > 0 {
		arg := fmt.Sprintf("messages=[%s]", strings.Join(sampleProto.TargetedMessages, messageDelimiter))
		codeGeneratorRequest.Parameter = &arg
	}

	// Perform the conversion:
	response, err := protoConverter.convert(&codeGeneratorRequest)
	require.NoError(t, err, "Unable to convert sample proto file (%v)", sampleProtoFileName)
	require.Equal(t, len(sampleProto.ExpectedJSONSchema), len(response.File), "Incorrect number of JSON-Schema files returned for sample proto file (%v)", sampleProtoFileName)

	for responseFileIndex, responseFile := range response.File {
		expected := strings.TrimSpace(sampleProto.ExpectedJSONSchema[responseFileIndex])
		got := responseFile.GetContent()

		if diff := cmp.Diff(expected, got); diff != "" {
			t.Error(diff)
		}

		// Validate the generated filenames:
		if len(sampleProto.ExpectedFileNames) > 0 {
			assert.Equal(t, sampleProto.ExpectedFileNames[responseFileIndex], responseFile.GetName())
		}

		// Validate any intended-to-fail data against the new schema:
		if len(sampleProto.ObjectsToValidateFail) >= responseFileIndex+1 {
			valid, err := validateSchema(got, sampleProto.ObjectsToValidateFail[responseFileIndex])
			assert.NoError(t, err)
			assert.False(t, valid, "Expected canned data to fail validation)")
		}

		// Validate any intended-to-pass data against the new schema:
		if len(sampleProto.ObjectsToValidatePass) >= responseFileIndex+1 {
			valid, err := validateSchema(got, sampleProto.ObjectsToValidatePass[responseFileIndex])
			assert.NoError(t, err, "Error validating canned data with generated schema")
			assert.True(t, valid, "Expected canned data validate)")
		}
	}

	// Return now if we have no files
	if len(response.File) == 0 {
		return
	}

	// Check for the correct prefix:
	if protoConverter.Flags.PrefixSchemaFilesWithPackage {
		assert.Contains(t, response.File[0].GetName(), "samples")
	} else {
		assert.NotContains(t, response.File[0].GetName(), "samples")
	}
}

func configureSampleProtos() map[string]sampleProto {
	return map[string]sampleProto{
		"AllRequired": {
			Flags:                 ConverterFlags{AllFieldsRequired: true},
			ExpectedJSONSchema:    []string{testdata.PayloadMessage2},
			FilesToGenerate:       []string{"PayloadMessage2.proto"},
			ProtoFileName:         "PayloadMessage2.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessage2Fail},
			ObjectsToValidatePass: []string{testdata.PayloadMessage2Pass},
		},
		"ArrayOfEnums": {
			ExpectedJSONSchema:    []string{testdata.ArrayOfEnums},
			FilesToGenerate:       []string{"ArrayOfEnums.proto"},
			ProtoFileName:         "ArrayOfEnums.proto",
			ObjectsToValidateFail: []string{testdata.ArrayOfEnumsFail},
			ObjectsToValidatePass: []string{testdata.ArrayOfEnumsPass},
		},
		"ArrayOfMessages": {
			ExpectedJSONSchema:    []string{testdata.PayloadMessage, testdata.ArrayOfMessages},
			FilesToGenerate:       []string{"ArrayOfMessages.proto", "PayloadMessage.proto"},
			ProtoFileName:         "ArrayOfMessages.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessageFail, testdata.ArrayOfMessagesFail},
			ObjectsToValidatePass: []string{testdata.PayloadMessagePass, testdata.ArrayOfMessagesPass},
		},
		"TypeNamesWithNoPackage": {
			Flags:                 ConverterFlags{TypeNamesWithNoPackage: true},
			ExpectedJSONSchema:    []string{testdata.PayloadMessage, testdata.TypeNamesWithNoPackage},
			FilesToGenerate:       []string{"ArrayOfMessages.proto", "PayloadMessage.proto"},
			ProtoFileName:         "ArrayOfMessages.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessageFail, testdata.TypeNamesWithNoPackageFail},
			ObjectsToValidatePass: []string{testdata.PayloadMessagePass, testdata.TypeNamesWithNoPackagePass},
		},
		"ArrayOfObjects": {
			Flags:                 ConverterFlags{AllowNullValues: true},
			ExpectedJSONSchema:    []string{testdata.ArrayOfObjects},
			FilesToGenerate:       []string{"ArrayOfObjects.proto"},
			ProtoFileName:         "ArrayOfObjects.proto",
			ObjectsToValidateFail: []string{testdata.ArrayOfObjectsFail},
			ObjectsToValidatePass: []string{testdata.ArrayOfObjectsPass},
		},
		"ArrayOfPrimitives": {
			Flags:                 ConverterFlags{AllowNullValues: true},
			ExpectedJSONSchema:    []string{testdata.ArrayOfPrimitives},
			FilesToGenerate:       []string{"ArrayOfPrimitives.proto"},
			ProtoFileName:         "ArrayOfPrimitives.proto",
			ObjectsToValidateFail: []string{testdata.ArrayOfPrimitivesFail},
			ObjectsToValidatePass: []string{testdata.ArrayOfPrimitivesPass},
		},
		"ArrayOfPrimitivesDouble": {
			Flags: ConverterFlags{
				AllowNullValues:           true,
				UseProtoAndJSONFieldNames: true,
			},
			ExpectedJSONSchema:    []string{testdata.ArrayOfPrimitivesDouble},
			FilesToGenerate:       []string{"ArrayOfPrimitives.proto"},
			ProtoFileName:         "ArrayOfPrimitives.proto",
			ObjectsToValidateFail: []string{testdata.ArrayOfPrimitivesDoubleFail},
			ObjectsToValidatePass: []string{testdata.ArrayOfPrimitivesDoublePass},
		},
		"BytesPayload": {
			ExpectedJSONSchema:    []string{testdata.BytesPayload},
			FilesToGenerate:       []string{"BytesPayload.proto"},
			ProtoFileName:         "BytesPayload.proto",
			ObjectsToValidateFail: []string{testdata.BytesPayloadFail},
		},
		"Comments": {
			ExpectedJSONSchema:    []string{testdata.MessageWithComments},
			FilesToGenerate:       []string{"MessageWithComments.proto"},
			ProtoFileName:         "MessageWithComments.proto",
			ObjectsToValidateFail: []string{testdata.MessageWithCommentsFail},
		},
		"CyclicalReference": {
			ExpectedJSONSchema: []string{testdata.CyclicalReferenceMessageM, testdata.CyclicalReferenceMessageFoo, testdata.CyclicalReferenceMessageBar, testdata.CyclicalReferenceMessageBaz},
			FilesToGenerate:    []string{"CyclicalReference.proto"},
			ProtoFileName:      "CyclicalReference.proto",
		},
		"EnumNestedReference": {
			ExpectedJSONSchema:    []string{testdata.EnumNestedReference},
			FilesToGenerate:       []string{"EnumNestedReference.proto"},
			ProtoFileName:         "EnumNestedReference.proto",
			ObjectsToValidateFail: []string{testdata.EnumNestedReferenceFail},
			ObjectsToValidatePass: []string{testdata.EnumNestedReferencePass},
		},
		"EnumWithMessage": {
			ExpectedJSONSchema:    []string{testdata.EnumWithMessage},
			FilesToGenerate:       []string{"EnumWithMessage.proto"},
			ProtoFileName:         "EnumWithMessage.proto",
			ObjectsToValidateFail: []string{testdata.EnumWithMessageFail},
			ObjectsToValidatePass: []string{testdata.EnumWithMessagePass},
		},
		"EnumCeption": {
			ExpectedJSONSchema:    []string{testdata.PayloadMessage, testdata.EnumCeption},
			FilesToGenerate:       []string{"Enumception.proto", "PayloadMessage.proto"},
			ProtoFileName:         "Enumception.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessageFail, testdata.EnumCeptionFail},
			ObjectsToValidatePass: []string{testdata.PayloadMessagePass, testdata.EnumCeptionPass},
		},
		"GoogleValue": {
			Flags:                 ConverterFlags{DisallowAdditionalProperties: true},
			ExpectedJSONSchema:    []string{testdata.GoogleValue},
			FilesToGenerate:       []string{"GoogleValue.proto"},
			ProtoFileName:         "GoogleValue.proto",
			ObjectsToValidateFail: []string{testdata.GoogleValueFail},
			ObjectsToValidatePass: []string{testdata.GoogleValuePass},
		},
		"GoogleInt64Value": {
			ExpectedJSONSchema:    []string{testdata.GoogleInt64Value},
			FilesToGenerate:       []string{"GoogleInt64Value.proto"},
			ProtoFileName:         "GoogleInt64Value.proto",
			ObjectsToValidateFail: []string{testdata.GoogleInt64ValueFail},
			ObjectsToValidatePass: []string{testdata.GoogleInt64ValuePass},
		},
		"GoogleInt64ValueAllowNull": {
			Flags:                 ConverterFlags{AllowNullValues: true},
			ExpectedJSONSchema:    []string{testdata.GoogleInt64ValueAllowNull},
			FilesToGenerate:       []string{"GoogleInt64ValueAllowNull.proto"},
			ProtoFileName:         "GoogleInt64ValueAllowNull.proto",
			ObjectsToValidateFail: []string{testdata.GoogleInt64ValueAllowNullFail},
			ObjectsToValidatePass: []string{testdata.GoogleInt64ValueAllowNullPass},
		},
		"GoogleInt64ValueDisallowString": {
			Flags:                 ConverterFlags{DisallowBigIntsAsStrings: true},
			ExpectedJSONSchema:    []string{testdata.GoogleInt64ValueDisallowString},
			FilesToGenerate:       []string{"GoogleInt64ValueDisallowString.proto"},
			ProtoFileName:         "GoogleInt64ValueDisallowString.proto",
			ObjectsToValidateFail: []string{testdata.GoogleInt64ValueDisallowStringFail},
			ObjectsToValidatePass: []string{testdata.GoogleInt64ValueDisallowStringPass},
		},
		"GoogleInt64ValueDisallowStringAllowNull": {
			Flags: ConverterFlags{
				DisallowBigIntsAsStrings: true,
				AllowNullValues:          true,
			},
			ExpectedJSONSchema:    []string{testdata.GoogleInt64ValueDisallowStringAllowNull},
			FilesToGenerate:       []string{"GoogleInt64ValueDisallowStringAllowNull.proto"},
			ProtoFileName:         "GoogleInt64ValueDisallowStringAllowNull.proto",
			ObjectsToValidateFail: []string{testdata.GoogleInt64ValueDisallowStringAllowNullFail},
			ObjectsToValidatePass: []string{testdata.GoogleInt64ValueDisallowStringAllowNullPass},
		},
		"JSONFields": {
			Flags:                 ConverterFlags{UseJSONFieldnamesOnly: true},
			ExpectedJSONSchema:    []string{testdata.JSONFields},
			FilesToGenerate:       []string{"JSONFields.proto"},
			ProtoFileName:         "JSONFields.proto",
			ObjectsToValidateFail: []string{testdata.JSONFieldsFail},
			ObjectsToValidatePass: []string{testdata.JSONFieldsPass},
		},
		"Maps": {
			ExpectedJSONSchema:    []string{testdata.Maps},
			FilesToGenerate:       []string{"Maps.proto"},
			ProtoFileName:         "Maps.proto",
			ObjectsToValidateFail: []string{testdata.MapsFail},
			ObjectsToValidatePass: []string{testdata.MapsPass},
		},
		"NestedMessage": {
			ExpectedJSONSchema:    []string{testdata.PayloadMessage, testdata.NestedMessage},
			FilesToGenerate:       []string{"NestedMessage.proto", "PayloadMessage.proto"},
			ProtoFileName:         "NestedMessage.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessageFail, testdata.NestedMessageFail},
			ObjectsToValidatePass: []string{testdata.PayloadMessagePass, testdata.NestedMessagePass},
		},
		"NestedObject": {
			ExpectedJSONSchema:    []string{testdata.NestedObject},
			FilesToGenerate:       []string{"NestedObject.proto"},
			ProtoFileName:         "NestedObject.proto",
			ObjectsToValidateFail: []string{testdata.NestedObjectFail},
			ObjectsToValidatePass: []string{testdata.NestedObjectPass},
		},
		"NoPackage": {
			ExpectedJSONSchema: []string{},
			FilesToGenerate:    []string{},
			ProtoFileName:      "NoPackage.proto",
		},
		"OneOf": {
			Flags:                 ConverterFlags{AllFieldsRequired: true, EnforceOneOf: true},
			ExpectedJSONSchema:    []string{testdata.OneOf},
			FilesToGenerate:       []string{"OneOf.proto"},
			ProtoFileName:         "OneOf.proto",
			ObjectsToValidateFail: []string{testdata.OneOfFail},
			ObjectsToValidatePass: []string{testdata.OneOfPass},
		},
		"PackagePrefix": {
			Flags:                 ConverterFlags{PrefixSchemaFilesWithPackage: true},
			ExpectedJSONSchema:    []string{testdata.Timestamp},
			FilesToGenerate:       []string{"Timestamp.proto"},
			ProtoFileName:         "Timestamp.proto",
			ObjectsToValidateFail: []string{testdata.TimestampFail},
			ObjectsToValidatePass: []string{testdata.TimestampPass},
		},
		"PayloadMessage": {
			ExpectedJSONSchema:    []string{testdata.PayloadMessage},
			FilesToGenerate:       []string{"PayloadMessage.proto"},
			ProtoFileName:         "PayloadMessage.proto",
			ObjectsToValidateFail: []string{testdata.PayloadMessageFail},
			ObjectsToValidatePass: []string{testdata.PayloadMessagePass},
		},
		"Proto2NestedMessage": {
			ExpectedJSONSchema:    []string{testdata.Proto2PayloadMessage, testdata.Proto2NestedMessage},
			FilesToGenerate:       []string{"Proto2PayloadMessage.proto", "Proto2NestedMessage.proto"},
			ProtoFileName:         "Proto2NestedMessage.proto",
			ObjectsToValidateFail: []string{testdata.Proto2PayloadMessageFail, testdata.Proto2NestedMessageFail},
			ObjectsToValidatePass: []string{testdata.Proto2PayloadMessagePass, testdata.Proto2NestedMessagePass},
		},
		"Proto2NestedObject": {
			Flags:                 ConverterFlags{AllFieldsRequired: true},
			ExpectedJSONSchema:    []string{testdata.Proto2NestedObject},
			FilesToGenerate:       []string{"Proto2NestedObject.proto"},
			ProtoFileName:         "Proto2NestedObject.proto",
			ObjectsToValidateFail: []string{testdata.Proto2NestedObjectFail},
			ObjectsToValidatePass: []string{testdata.Proto2NestedObjectPass},
		},
		"Proto2Required": {
			ExpectedJSONSchema:    []string{testdata.Proto2Required},
			FilesToGenerate:       []string{"Proto2Required.proto"},
			ProtoFileName:         "Proto2Required.proto",
			ObjectsToValidateFail: []string{testdata.Proto2RequiredFail},
			ObjectsToValidatePass: []string{testdata.Proto2RequiredPass},
		},
		"SelfReference": {
			ExpectedJSONSchema:    []string{testdata.SelfReference},
			FilesToGenerate:       []string{"SelfReference.proto"},
			ProtoFileName:         "SelfReference.proto",
			ObjectsToValidateFail: []string{testdata.SelfReferenceFail},
			ObjectsToValidatePass: []string{testdata.SelfReferencePass},
		},
		"SeveralEnums": {
			ExpectedJSONSchema:    []string{testdata.FirstEnum, testdata.SecondEnum},
			FilesToGenerate:       []string{"SeveralEnums.proto"},
			ProtoFileName:         "SeveralEnums.proto",
			ObjectsToValidateFail: []string{testdata.FirstEnumFail, testdata.SecondEnumFail},
			ObjectsToValidatePass: []string{testdata.FirstEnumPass, testdata.SecondEnumPass},
		},
		"SeveralMessages": {
			ExpectedJSONSchema:    []string{testdata.FirstMessage, testdata.SecondMessage},
			FilesToGenerate:       []string{"SeveralMessages.proto"},
			ProtoFileName:         "SeveralMessages.proto",
			ObjectsToValidateFail: []string{testdata.FirstMessageFail, testdata.SecondMessageFail},
			ObjectsToValidatePass: []string{testdata.FirstMessagePass, testdata.SecondMessagePass},
		},
		"TargetedMessages": {
			TargetedMessages:   []string{"MessageKind10", "MessageKind11", "MessageKind12"},
			ExpectedJSONSchema: []string{testdata.MessageKind10, testdata.MessageKind11, testdata.MessageKind12},
			FilesToGenerate:    []string{"TwelveMessages.proto"},
			ProtoFileName:      "TwelveMessages.proto",
		},
		"Timestamp": {
			ExpectedJSONSchema:    []string{testdata.Timestamp},
			FilesToGenerate:       []string{"Timestamp.proto"},
			ProtoFileName:         "Timestamp.proto",
			ObjectsToValidateFail: []string{testdata.TimestampFail},
			ObjectsToValidatePass: []string{testdata.TimestampPass},
		},
		"WellKnown": {
			ExpectedJSONSchema:    []string{testdata.WellKnown},
			FilesToGenerate:       []string{"WellKnown.proto"},
			ProtoFileName:         "WellKnown.proto",
			ObjectsToValidateFail: []string{testdata.WellKnownFail},
			ObjectsToValidatePass: []string{testdata.WellKnownPass},
		},
	}
}

// Load the specified .proto files into a FileDescriptorSet. Any errors in loading/parsing will
// immediately fail the test.
func mustReadProtoFiles(t *testing.T, includePath string, filenames ...string) *descriptorpb.FileDescriptorSet {
	protocBinary, err := exec.LookPath("protoc")
	if err != nil {
		t.Fatalf("Can't find 'protoc' binary in $PATH: %s", err.Error())
	}

	// Use protoc to output descriptorpb info for the specified .proto files.
	var args []string
	args = append(args, "--descriptor_set_out=/dev/stdout")
	args = append(args, "--include_source_info")
	args = append(args, "--include_imports")
	args = append(args, "-I../../")
	args = append(args, "--proto_path="+includePath)
	args = append(args, filenames...)
	cmd := exec.Command(protocBinary, args...)
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to load descriptorpb set (%s): %s: %s",
			strings.Join(cmd.Args, " "), err.Error(), stderrBuf.String())
	}
	fds := &descriptorpb.FileDescriptorSet{}
	err = proto.Unmarshal(stdoutBuf.Bytes(), fds)
	if err != nil {
		t.Fatalf("failed to parse protoc output as FileDescriptorSet: %s", err.Error())
	}
	return fds
}

func validateSchema(jsonSchema, jsonData string) (bool, error) {
	var valid = false

	// Load the JSON schema:
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)

	// Load the JSON document we'll be validating:
	documentLoader := gojsonschema.NewStringLoader(jsonData)

	// Validate:
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil || result == nil {
		return valid, err
	}

	return result.Valid(), nil
}
