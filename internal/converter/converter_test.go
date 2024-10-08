package converter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/sixt/protoc-gen-jsonschema/internal/converter/testdata"
)

var (
	sampleProtoDirectory = "testdata/proto"
	sampleProtos         = make(map[string]sampleProto)
)

type sampleProto struct {
	AllowNullValues           bool
	ExpectedJSONSchema        []string
	FilesToGenerate           []string
	ProtoFileName             string
	UseProtoAndJSONFieldNames bool
}

func TestGenerateJsonSchema(t *testing.T) {

	// Configure the list of sample protos to test, and their expected JSON-Schemas:
	configureSampleProtos()

	// Convert the protos, compare the results against the expected JSON-Schemas:
	testConvertSampleProto(t, sampleProtos["Comments"])
	testConvertSampleProto(t, sampleProtos["ArrayOfMessages"])
	testConvertSampleProto(t, sampleProtos["ArrayOfObjects"])
	testConvertSampleProto(t, sampleProtos["ArrayOfPrimitives"])
	testConvertSampleProto(t, sampleProtos["ArrayOfPrimitivesDouble"])
	testConvertSampleProto(t, sampleProtos["EnumCeption"])
	testConvertSampleProto(t, sampleProtos["ImportedEnum"])
	testConvertSampleProto(t, sampleProtos["NestedMessage"])
	testConvertSampleProto(t, sampleProtos["NestedObject"])
	testConvertSampleProto(t, sampleProtos["PayloadMessage"])
	testConvertSampleProto(t, sampleProtos["SeveralEnums"])
	testConvertSampleProto(t, sampleProtos["SeveralMessages"])
	testConvertSampleProto(t, sampleProtos["ArrayOfEnums"])
	testConvertSampleProto(t, sampleProtos["Maps"])
	testConvertSampleProto(t, sampleProtos["WellKnown"])
}

func testConvertSampleProto(t *testing.T, sampleProto sampleProto) {

	// Make a Logrus logger:
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.SetOutput(os.Stderr)

	// Use the logger to make a Converter:
	protoConverter := New(logger)
	protoConverter.AllowNullValues = sampleProto.AllowNullValues
	protoConverter.UseProtoAndJSONFieldnames = sampleProto.UseProtoAndJSONFieldNames

	// Open the sample proto file:
	sampleProtoFileName := fmt.Sprintf("%v/%v", sampleProtoDirectory, sampleProto.ProtoFileName)
	fileDescriptorSet := mustReadProtoFiles(t, sampleProtoDirectory, sampleProto.ProtoFileName)

	// Prepare a request:
	codeGeneratorRequest := plugin.CodeGeneratorRequest{
		FileToGenerate: sampleProto.FilesToGenerate,
		ProtoFile:      fileDescriptorSet.GetFile(),
	}

	// Perform the conversion:
	response, err := protoConverter.convert(&codeGeneratorRequest)
	if err != nil {
		t.Fatal(err)
	}

	if len(response.File) != len(sampleProto.ExpectedJSONSchema) {
		t.Fatalf("Incorrect number of JSON-Schema files returned for sample proto file (%v)", sampleProtoFileName)
	}

	for i, file := range response.File {
		want := sampleProto.ExpectedJSONSchema[i]

		if diff := cmp.Diff(file.GetContent(), want); diff != "" {
			t.Errorf("differences: %s\n%s", file.GetName(), diff)
		}
	}
}

func configureSampleProtos() {
	// ArrayOfMessages:
	sampleProtos["ArrayOfMessages"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.ArrayOfMessages},
		FilesToGenerate:    []string{"ArrayOfMessages.proto", "PayloadMessage.proto"},
		ProtoFileName:      "ArrayOfMessages.proto",
	}

	// ArrayOfObjects:
	sampleProtos["ArrayOfObjects"] = sampleProto{
		AllowNullValues:    true,
		ExpectedJSONSchema: []string{testdata.ArrayOfObjects},
		FilesToGenerate:    []string{"ArrayOfObjects.proto"},
		ProtoFileName:      "ArrayOfObjects.proto",
	}

	// ArrayOfPrimitives:
	sampleProtos["ArrayOfPrimitives"] = sampleProto{
		AllowNullValues:    true,
		ExpectedJSONSchema: []string{testdata.ArrayOfPrimitives},
		FilesToGenerate:    []string{"ArrayOfPrimitives.proto"},
		ProtoFileName:      "ArrayOfPrimitives.proto",
	}

	// ArrayOfPrimitives:
	sampleProtos["ArrayOfPrimitivesDouble"] = sampleProto{
		AllowNullValues:           true,
		ExpectedJSONSchema:        []string{testdata.ArrayOfPrimitivesDouble},
		FilesToGenerate:           []string{"ArrayOfPrimitives.proto"},
		ProtoFileName:             "ArrayOfPrimitives.proto",
		UseProtoAndJSONFieldNames: true,
	}

	// EnumCeption:
	sampleProtos["EnumCeption"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.ImportedEnum, testdata.EnumCeption},
		FilesToGenerate:    []string{"Enumception.proto", "PayloadMessage.proto", "ImportedEnum.proto"},
		ProtoFileName:      "Enumception.proto",
	}

	// ImportedEnum:
	sampleProtos["ImportedEnum"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.ImportedEnum},
		FilesToGenerate:    []string{"ImportedEnum.proto"},
		ProtoFileName:      "ImportedEnum.proto",
	}

	// NestedMessage:
	sampleProtos["NestedMessage"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.NestedMessage},
		FilesToGenerate:    []string{"NestedMessage.proto", "PayloadMessage.proto"},
		ProtoFileName:      "NestedMessage.proto",
	}

	// NestedObject:
	sampleProtos["NestedObject"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.NestedObject},
		FilesToGenerate:    []string{"NestedObject.proto"},
		ProtoFileName:      "NestedObject.proto",
	}

	// PayloadMessage:
	sampleProtos["PayloadMessage"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.PayloadMessage},
		FilesToGenerate:    []string{"PayloadMessage.proto"},
		ProtoFileName:      "PayloadMessage.proto",
	}

	// SeveralEnums:
	sampleProtos["SeveralEnums"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.FirstEnum, testdata.SecondEnum},
		FilesToGenerate:    []string{"SeveralEnums.proto"},
		ProtoFileName:      "SeveralEnums.proto",
	}

	// SeveralMessages:
	sampleProtos["SeveralMessages"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.FirstMessage, testdata.SecondMessage},
		FilesToGenerate:    []string{"SeveralMessages.proto"},
		ProtoFileName:      "SeveralMessages.proto",
	}

	// ArrayOfEnums:
	sampleProtos["ArrayOfEnums"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.ArrayOfEnums},
		FilesToGenerate:    []string{"ArrayOfEnums.proto"},
		ProtoFileName:      "ArrayOfEnums.proto",
	}

	// Maps:
	sampleProtos["Maps"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.Maps},
		FilesToGenerate:    []string{"Maps.proto"},
		ProtoFileName:      "Maps.proto",
	}

	// Comments:
	sampleProtos["Comments"] = sampleProto{
		AllowNullValues:    false,
		ExpectedJSONSchema: []string{testdata.MessageWithComments},
		FilesToGenerate:    []string{"MessageWithComments.proto"},
		ProtoFileName:      "MessageWithComments.proto",
	}

	sampleProtos["WellKnown"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.WellKnown},
		FilesToGenerate:    []string{"WellKnown.proto"},
		ProtoFileName:      "WellKnown.proto",
	}
}

// Load the specified .proto files into a FileDescriptorSet. Any errors in loading/parsing will
// immediately fail the test.
func mustReadProtoFiles(t *testing.T, includePath string, filenames ...string) *descriptor.FileDescriptorSet {
	protocBinary, err := exec.LookPath("protoc")
	if err != nil {
		t.Fatalf("Can't find 'protoc' binary in $PATH: %s", err.Error())
	}

	// Use protoc to output descriptor info for the specified .proto files.
	var args []string
	args = append(args, "--descriptor_set_out=/dev/stdout")
	args = append(args, "--include_source_info")
	args = append(args, "--include_imports")
	args = append(args, "--proto_path="+includePath)
	args = append(args, filenames...)
	cmd := exec.Command(protocBinary, args...)
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to load descriptor set (%s): %s: %s",
			strings.Join(cmd.Args, " "), err.Error(), stderrBuf.String())
	}
	fds := &descriptor.FileDescriptorSet{}
	err = proto.Unmarshal(stdoutBuf.Bytes(), fds)
	if err != nil {
		t.Fatalf("failed to parse protoc output as FileDescriptorSet: %s", err.Error())
	}
	return fds
}
