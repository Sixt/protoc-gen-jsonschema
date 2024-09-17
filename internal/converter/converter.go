package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	"github.com/alecthomas/jsonschema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"github.com/sirupsen/logrus"
)

// Converter is everything you need to convert protos to JSONSchemas:
type Converter struct {
	DisallowAdditionalProperties bool
	UseProtoAndJSONFieldnames    bool
	Logger                       *logrus.Logger
	sourceInfo                   *sourceCodeInfo
}

// ConvertFrom tells the convert to work on the given input:
func (c *Converter) ConvertFrom(rd io.Reader) (*pluginpb.CodeGeneratorResponse, error) {
	c.Logger.Debug("Reading code generation request")

	input, err := ioutil.ReadAll(rd)
	if err != nil {
		c.Logger.WithError(err).Error("Failed to read request")
		return nil, err
	}

	req := new(pluginpb.CodeGeneratorRequest)
	err = proto.Unmarshal(input, req)
	if err != nil {
		c.Logger.WithError(err).Error("Can't unmarshal input")
		return nil, err
	}

	c.parseGeneratorParameters(req.GetParameter())

	c.Logger.Debug("Converting input")
	return c.convert(req)
	// return c.debugger(req)
}

func (c *Converter) parseGeneratorParameters(parameters string) {

	for _, parameter := range strings.Split(parameters, ",") {
		switch parameter {
		case "allow_null_values":
			// short-circuited to on
		case "debug":
			c.Logger.SetLevel(logrus.DebugLevel)
		case "disallow_additional_properties":
			c.DisallowAdditionalProperties = true
		case "disallow_bigints_as_strings":
			// short-circuited to off
		case "proto_and_json_fieldnames":
			c.UseProtoAndJSONFieldnames = true
		}
	}
}

// Converts a proto "ENUM" into a JSON-Schema:
func (c *Converter) convertEnumType(enum *descriptorpb.EnumDescriptorProto) (jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	jsonSchemaType := jsonschema.Type{
		Version: jsonschema.Version,
	}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetEnum(enum); src != nil {
		jsonSchemaType.Description = formatDescription(src)
	}

	// Allow both strings and integers:
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "string"})
	jsonSchemaType.OneOf = append(jsonSchemaType.OneOf, &jsonschema.Type{Type: "integer"})

	// Add the allowed values:
	for _, enumValue := range enum.Value {
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Name)
		jsonSchemaType.Enum = append(jsonSchemaType.Enum, enumValue.Number)
	}

	return jsonSchemaType, nil
}

// Converts a proto file into a JSON-Schema:
func (c *Converter) convertFile(file *descriptorpb.FileDescriptorProto) ([]*pluginpb.CodeGeneratorResponse_File, error) {

	// Input filename:
	protoFilename := path.Base(file.GetName())
	logger := c.Logger.WithField("proto_filename", protoFilename)

	// Prepare a list of responses:
	var response []*pluginpb.CodeGeneratorResponse_File

	// Warn about multiple messages / enums in files:
	if len(file.GetMessageType()) > 1 {
		c.Logger.WithField("schemas", len(file.GetMessageType())).Warn("protoc-gen-jsonschema will create multiple MESSAGE schemas from one proto file")
	}

	if len(file.GetEnumType()) > 1 {
		c.Logger.WithField("schemas", len(file.GetMessageType())).Warn("protoc-gen-jsonschema will create multiple ENUM schemas from one proto file")
	}

	// Generate standalone ENUMs:
	if len(file.GetMessageType()) == 0 {
		for _, enum := range file.GetEnumType() {
			jsonFilename := fmt.Sprintf("%s.jsonschema", enum.GetName())

			logger := logger.WithFields(logrus.Fields{
				"enum_name": enum.GetName(),
				"jsonschema_filename": jsonFilename,
			})

			logger.Info("Generating JSON-schema for stand-alone ENUM")

			// Convert the ENUM:
			enumJSONSchema, err := c.convertEnumType(enum)
			if err != nil {
				logger.WithError(err).Error("Failed to convert")
				return nil, err
			}

			// Marshal the JSON-Schema into JSON:
			data, err := json.MarshalIndent(enumJSONSchema, "", "    ")
			if err != nil {
				c.Logger.WithError(err).Error("Failed to encode jsonSchema")
				return nil, err
			}

			// Add a response:
			response = append(response, &pluginpb.CodeGeneratorResponse_File{
				Name:    proto.String(jsonFilename),
				Content: proto.String(string(data)),
			})
		}

		return response, nil
	}

	// Otherwise process MESSAGES (packages):
	pkg, ok := c.relativelyLookupPackage(globalPkg, file.GetPackage())
	if !ok {
		return nil, fmt.Errorf("no such package found: %s", file.GetPackage())
	}

	for _, msg := range file.GetMessageType() {
		jsonFilename := fmt.Sprintf("%s.jsonschema", msg.GetName())

		logger := logger.WithFields(logrus.Fields{
			"msg_name": msg.GetName(),
			"jsonschema_filename": jsonFilename,
		})
		logger.Info("Generating JSON-schema for MESSAGE")

		// Convert the message:
		msgSchema, err := c.convertMessageType(pkg, msg, "")
		if err != nil {
			logger.WithError(err).Error("Failed to convert")
			return nil, err
		}

		// Marshal the JSON-Schema into JSON:
		data, err := json.MarshalIndent(msgSchema, "", "    ")
		if err != nil {
			logger.WithError(err).Error("Failed to encode jsonschema")
			return nil, err
		}

		// Add a response:
		response = append(response, &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(jsonFilename),
			Content: proto.String(string(data)),
		})
	}

	return response, nil
}

func (c *Converter) convert(req *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	generateTargets := make(map[string]bool)
	for _, file := range req.GetFileToGenerate() {
		generateTargets[file] = true
	}

	c.sourceInfo = newSourceCodeInfo(req.GetProtoFile())
	res := &pluginpb.CodeGeneratorResponse{}
	for _, file := range req.GetProtoFile() {
		for _, msg := range file.GetMessageType() {
			c.Logger.WithField("msg_name", msg.GetName()).WithField("package_name", file.GetPackage()).Debug("Loading a message")
			c.registerType(file.GetPackage(), msg)
		}
	}
	for _, file := range req.GetProtoFile() {
		if _, ok := generateTargets[file.GetName()]; ok {
			c.Logger.WithField("filename", file.GetName()).Debug("Converting file")
			converted, err := c.convertFile(file)
			if err != nil {
				res.Error = proto.String(fmt.Sprintf("Failed to convert %s: %v", file.GetName(), err))
				return res, err
			}
			res.File = append(res.File, converted...)
		}
	}
	return res, nil
}
