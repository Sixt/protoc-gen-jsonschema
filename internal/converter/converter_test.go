package converter

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"github.com/sirupsen/logrus"
	"github.com/sixt/protoc-gen-jsonschema/internal/converter/testdata"
)

var (
	sampleProtoDirectory = "testdata/proto"
	sampleProtos         = make(map[string]sampleProto)
)

type sampleProto struct {
	ExpectedJSONSchema        []string
	FilesToGenerate           []string
	ProtoFileName             string
	UseProtoAndJSONFieldNames bool
	DisallowAdditionalProperties bool
}

func TestGenerateJsonSchema(t *testing.T) {

	// Configure the list of sample protos to test, and their expected JSON-Schemas:
	configureSampleProtos()

	// Convert the protos, compare the results against the expected JSON-Schemas:
	testConvertSampleProto(t, "Comments")
	testConvertSampleProto(t, "ArrayOfMessages")
	testConvertSampleProto(t, "ArrayOfObjects")
	testConvertSampleProto(t, "ArrayOfPrimitives")
	testConvertSampleProto(t, "ArrayOfPrimitivesDouble")
	testConvertSampleProto(t, "Enumception")
	testConvertSampleProto(t, "ImportedEnum")
	testConvertSampleProto(t, "NestedMessage")
	testConvertSampleProto(t, "NestedObject")
	testConvertSampleProto(t, "PayloadMessage")
	testConvertSampleProto(t, "SeveralEnums")
	testConvertSampleProto(t, "SeveralMessages")
	testConvertSampleProto(t, "ArrayOfEnums")
	testConvertSampleProto(t, "Maps")
	testConvertSampleProto(t, "WellKnown")
}

func testConvertSampleProto(t *testing.T, name string) {

	sampleProto := sampleProtos[name]
	t.Log("running test:", name)

	// Make a Logrus logger:
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.SetOutput(os.Stderr)

	// Use the logger to make a Converter:
	protoConverter := &Converter{
		Logger: logger,
	}
	protoConverter.UseProtoAndJSONFieldnames = sampleProto.UseProtoAndJSONFieldNames
	protoConverter.DisallowAdditionalProperties = sampleProto.DisallowAdditionalProperties

	// Open the sample proto file:
	sampleProtoFileName := sampleProtoDirectory + "/" + sampleProto.ProtoFileName
	fileDescriptorSet := mustReadProtoFiles(t, sampleProtoDirectory, sampleProto.ProtoFileName)

	// Prepare a request:
	codeGeneratorRequest := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: sampleProto.FilesToGenerate,
		ProtoFile:      fileDescriptorSet.GetFile(),
	}

	// Perform the conversion:
	response, err := protoConverter.convert(codeGeneratorRequest)
	if err != nil {
		t.Fatal(err)
	}

	if len(response.File) != len(sampleProto.ExpectedJSONSchema) {
		t.Fatal("Incorrect number of JSON-Schema files returned for sample proto file:", sampleProtoFileName)
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
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.ArrayOfMessages},
		FilesToGenerate:    []string{"ArrayOfMessages.proto", "PayloadMessage.proto"},
		ProtoFileName:      "ArrayOfMessages.proto",
	}

	// ArrayOfObjects:
	sampleProtos["ArrayOfObjects"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.ArrayOfObjects},
		FilesToGenerate:    []string{"ArrayOfObjects.proto"},
		ProtoFileName:      "ArrayOfObjects.proto",
	}

	// ArrayOfPrimitives:
	sampleProtos["ArrayOfPrimitives"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.ArrayOfPrimitives},
		FilesToGenerate:    []string{"ArrayOfPrimitives.proto"},
		ProtoFileName:      "ArrayOfPrimitives.proto",
	}

	// ArrayOfPrimitives:
	sampleProtos["ArrayOfPrimitivesDouble"] = sampleProto{
		ExpectedJSONSchema:        []string{testdata.ArrayOfPrimitivesDouble},
		FilesToGenerate:           []string{"ArrayOfPrimitives.proto"},
		ProtoFileName:             "ArrayOfPrimitives.proto",

		UseProtoAndJSONFieldNames: true,
	}

	// Enumception:
	sampleProtos["Enumception"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.ImportedEnum, testdata.Enumception},
		FilesToGenerate:    []string{"Enumception.proto", "PayloadMessage.proto", "ImportedEnum.proto"},
		ProtoFileName:      "Enumception.proto",
	}

	// ImportedEnum:
	sampleProtos["ImportedEnum"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.ImportedEnum},
		FilesToGenerate:    []string{"ImportedEnum.proto"},
		ProtoFileName:      "ImportedEnum.proto",
	}

	// NestedMessage:
	sampleProtos["NestedMessage"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.PayloadMessage, testdata.NestedMessage},
		FilesToGenerate:    []string{"NestedMessage.proto", "PayloadMessage.proto"},
		ProtoFileName:      "NestedMessage.proto",

	}

	// NestedObject:
	sampleProtos["NestedObject"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.NestedObject},
		FilesToGenerate:    []string{"NestedObject.proto"},
		ProtoFileName:      "NestedObject.proto",
	}

	// PayloadMessage:
	sampleProtos["PayloadMessage"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.PayloadMessage},
		FilesToGenerate:    []string{"PayloadMessage.proto"},
		ProtoFileName:      "PayloadMessage.proto",
	}

	// SeveralEnums:
	sampleProtos["SeveralEnums"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.FirstEnum, testdata.SecondEnum},
		FilesToGenerate:    []string{"SeveralEnums.proto"},
		ProtoFileName:      "SeveralEnums.proto",
	}

	// SeveralMessages:
	sampleProtos["SeveralMessages"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.FirstMessage, testdata.SecondMessage},
		FilesToGenerate:    []string{"SeveralMessages.proto"},
		ProtoFileName:      "SeveralMessages.proto",
	}

	// ArrayOfEnums:
	sampleProtos["ArrayOfEnums"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.ArrayOfEnums},
		FilesToGenerate:    []string{"ArrayOfEnums.proto"},
		ProtoFileName:      "ArrayOfEnums.proto",
	}

	// Maps:
	sampleProtos["Maps"] = sampleProto{
		ExpectedJSONSchema: []string{testdata.Maps},
		FilesToGenerate:    []string{"Maps.proto"},
		ProtoFileName:      "Maps.proto",
	}

	// Comments:
	sampleProtos["Comments"] = sampleProto{
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
func mustReadProtoFiles(t *testing.T, includePath string, filenames ...string) *descriptorpb.FileDescriptorSet {
	protocBinary, err := exec.LookPath("protoc")
	if err != nil {
		t.Fatal("Can't find 'protoc' binary in $PATH:", err)
	}

	// Use protoc to output descriptorpb info for the specified .proto files.
	args := []string{
		"--descriptor_set_out=/dev/stdout",
		"--include_source_info",
		"--include_imports",
		"--proto_path="+includePath,
	}
	args = append(args, filenames...)

	cmd := exec.Command(protocBinary, args...)
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		t.Fatalf("failed to load descriptorpb set (%s): %s: %s", strings.Join(cmd.Args, " "), err, stderrBuf.String())
	}

	fds := new(descriptorpb.FileDescriptorSet)
	if err := proto.Unmarshal(stdoutBuf.Bytes(), fds); err != nil {
		t.Fatal("failed to parse protoc output as FileDescriptorSet:", err)
	}

	return fds
}
