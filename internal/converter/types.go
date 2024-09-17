package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/alecthomas/jsonschema"
	"google.golang.org/protobuf/types/descriptorpb"
	"github.com/iancoleman/orderedmap"
	"github.com/xeipuuv/gojsonschema"
)

var globalPkg = &ProtoPackage{
	children: make(map[string]*ProtoPackage),
	types:    make(map[string]*descriptorpb.DescriptorProto),
}

var wellKnownTypes = map[string]string{
	"BoolValue":   gojsonschema.TYPE_BOOLEAN,

	// We _intentionally_ do not accept "NaN" and "Infinity".
	// These usually reflect a bug in the generating code, and thus a useful signal.
	"DoubleValue": gojsonschema.TYPE_NUMBER,
	"FloatValue":  gojsonschema.TYPE_NUMBER,

	"ListValue":   gojsonschema.TYPE_ARRAY,
	"NullValue":   gojsonschema.TYPE_NULL,
	"StringValue": gojsonschema.TYPE_STRING,
	"Struct":      gojsonschema.TYPE_OBJECT,
}

var (
	IntString = &jsonschema.Type{
		Type: gojsonschema.TYPE_STRING,
		Format: "regex",
		Pattern: `^-?[0-9]+$`,
	}
	UIntString = &jsonschema.Type{
		Type: gojsonschema.TYPE_STRING,
		Format: "regex",
		Pattern: `^?[0-9]+$`,
	}
)


func (c *Converter) registerType(pkgName string, msg *descriptorpb.DescriptorProto) {
	pkg := globalPkg

	for _, node := range strings.Split(pkgName, ".") {
		if pkg == globalPkg && node == "" {
			// Skips leading "."
			continue
		}

		child, ok := pkg.children[node]
		if !ok {
			child = &ProtoPackage{
				name:     pkg.name + "." + node,
				parent:   pkg,
				types:    make(map[string]*descriptorpb.DescriptorProto),
			}

			if pkg.children == nil {
				pkg.children = make(map[string]*ProtoPackage)
			}
			pkg.children[node] = child
		}

		pkg = child
	}

	pkg.types[msg.GetName()] = msg
}

func (c *Converter) relativelyLookupNestedType(desc *descriptorpb.DescriptorProto, name string) (*descriptorpb.DescriptorProto, bool) {
componentLoop:
	for _, component := range strings.Split(name, ".") {
		for _, nested := range desc.GetNestedType() {
			if nested.GetName() == component {
				desc = nested
				continue componentLoop
			}
		}

		c.Logger.WithField("component", component).WithField("description", desc.GetName()).Info("no such nested message")
		return nil, false
	}

	return desc, true
}

// Convert a proto "field" (essentially a type-switch with some recursion):
func (c *Converter) convertField(curPkg *ProtoPackage, desc *descriptorpb.FieldDescriptorProto, msg *descriptorpb.DescriptorProto) (*jsonschema.Type, error) {

	// Prepare a new jsonschema.Type for our eventual return value:
	var typ jsonschema.Type

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetField(desc); src != nil {
		typ.Description = formatDescription(src)
	}

	// Switch the types, and pick a JSONSchema equivalent:
	switch desc.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		if true {
			typ.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_BOOLEAN},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			typ.Type = gojsonschema.TYPE_BOOLEAN
		}

	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE, descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		if true {
			typ.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_NUMBER},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			typ.Type = gojsonschema.TYPE_NUMBER
		}

	case descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_UINT32,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED32, descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
		descriptorpb.FieldDescriptorProto_TYPE_INT32:
		if true {
			typ.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_INTEGER},
				IntString,
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			typ.Type = gojsonschema.TYPE_INTEGER
		}

	case descriptorpb.FieldDescriptorProto_TYPE_SINT64, descriptorpb.FieldDescriptorProto_TYPE_UINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
		descriptorpb.FieldDescriptorProto_TYPE_INT64:

		typ.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_STRING},
			UIntString,
			{Type: gojsonschema.TYPE_INTEGER},
		}

		if true {
			typ.OneOf = append(typ.OneOf,
				&jsonschema.Type{Type: gojsonschema.TYPE_NULL},
			)
		}

	case descriptorpb.FieldDescriptorProto_TYPE_STRING,
		descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		if true {
			typ.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_STRING},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			typ.Type = gojsonschema.TYPE_STRING
		}

	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		if strings.HasPrefix(desc.GetTypeName(), ".google.protobuf.") {
			switch desc.GetTypeName() {
			case ".google.protobuf.NullValue":
				return &jsonschema.Type{Type: gojsonschema.TYPE_NULL}, nil
			}
		}

		c.Logger.WithFields(logrus.Fields{
			"name":     desc.GetTypeName(),
		}).Error("enum value")

		typ.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_STRING},
			{Type: gojsonschema.TYPE_INTEGER},
		}

		if true {
			typ.OneOf = append(typ.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
		}

		name := msg.GetName()
		typName := desc.GetTypeName()
		
		// Go through all the enums we have, see if we can match any to this field by name:
		for _, enumDesc := range msg.GetEnumType() {

			// Each one has several values:
			for _, enumVal := range enumDesc.Value {

				// Figure out the entire name of this field:
				fullFieldName := fmt.Sprintf(".%v.%v", name, enumDesc.GetName())

				// If we find ENUM values for this field then put them into the JSONSchema list of allowed ENUM values:
				if strings.HasSuffix(typName, fullFieldName) {
					typ.Enum = append(typ.Enum, enumVal.Name)
					typ.Enum = append(typ.Enum, enumVal.Number)
				}
			}
		}

	case descriptorpb.FieldDescriptorProto_TYPE_GROUP, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		typ.Type = gojsonschema.TYPE_OBJECT

		if desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL {
			typ.AdditionalProperties = []byte("true")
		}

		if desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED {
			typ.AdditionalProperties = []byte("false")
		}

	default:
		return nil, fmt.Errorf("unrecognized field type: %s", desc.GetType())
	}

	// Recurse array of primitive types:
	if desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED && typ.Type != gojsonschema.TYPE_OBJECT {
		typ.Items = new(jsonschema.Type)

		if len(typ.Enum) > 0 {
			typ.Items.Enum = typ.Enum
			typ.Items.OneOf = nil
			typ.Enum = nil
		} else {
			typ.Items.Type = typ.Type
			typ.Items.OneOf = typ.OneOf
		}

		if true {
			typ.OneOf = []*jsonschema.Type{
				{Type: gojsonschema.TYPE_ARRAY},
				{Type: gojsonschema.TYPE_NULL},
			}
		} else {
			typ.Type = gojsonschema.TYPE_ARRAY
		}

		return &typ, nil
	}

	// Recurse nested objects / arrays of objects (if necessary):
	if typ.Type == gojsonschema.TYPE_OBJECT {

		recordType, pkgName, ok := c.lookupType(curPkg, desc.GetTypeName())
		if !ok {
			return nil, fmt.Errorf("no such message type: %s", desc.GetTypeName())
		}

		// Recurse the recordType:
		recursedTyp, err := c.convertMessageType(curPkg, recordType, pkgName)
		if err != nil {
			return nil, err
		}

		// Maps, arrays, and objects are structured in different ways:
		switch {

		// Maps:
		case recordType.Options.GetMapEntry():
			c.Logger.WithFields(logrus.Fields{
				"field_name": recordType.GetName(),
				"msg_name":   msg.GetName(),
			}).Tracef("Is a map")

			// Make sure we have a "value":
			if recursedTyp.Properties == nil {
				return nil, fmt.Errorf("Unable to find 'value' property of MAP type")
			}

			value, ok := recursedTyp.Properties.Get("value")
			if !ok {
				return nil, fmt.Errorf("Unable to find 'value' property of MAP type")
			}

			// Marshal the "value" properties to JSON (because that's how we can pass on AdditionalProperties):
			data, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			typ.AdditionalProperties = data

		// Arrays:
		case desc.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED:
			typ.Items = recursedTyp
			typ.Type = gojsonschema.TYPE_ARRAY

		case msg.GetName() != "" && pkgName == ".google.protobuf":
			return recursedTyp, nil

		// Objects:
		default:
			isPrimitive := true

			for _, t := range typ.OneOf {
				if t.Type == gojsonschema.TYPE_OBJECT || t.Type == gojsonschema.TYPE_ARRAY {
					isPrimitive = false
				}
			}

			if len(typ.OneOf) != 0 && isPrimitive {
				typ.OneOf = recursedTyp.OneOf
				typ.AdditionalProperties = nil
				typ.Type = ""

			} else {
				typ.Properties = recursedTyp.Properties
			}
		}

		// Optionally allow NULL values, if not already nullable
		if true && len(typ.OneOf) == 0 {
			typ.OneOf = []*jsonschema.Type{
				{Type: typ.Type},
				{Type: gojsonschema.TYPE_NULL},
			}
			typ.Type = ""
		}
	}

	return &typ, nil
}

func convertWKT(name string) (*jsonschema.Type, error) {

	if typ := wellKnownTypes[name]; typ != "" {
		return &jsonschema.Type{Type: typ}, nil
	}

	switch name {
	case "BytesValue":
		return &jsonschema.Type{
			Type: gojsonschema.TYPE_STRING,
			Format: "regex",
			// must accept standard or URL-safe base64 with or without padding.
			Pattern: `^[-_A-Za-z0-9+\/]*={0,2}$`,
		}, nil

	case "Duration":
		return &jsonschema.Type{
			Type: gojsonschema.TYPE_STRING,
			Format: "regex",
			// must accept any fractional digits (also none) as long as they fit into nano-seconds.
			// the suffix "s" is required.
			Pattern: `^[0-9]+(\.[0-9]{0,9})?s$`,
		}, nil

	case "Empty":
		return &jsonschema.Type{
			Type:          gojsonschema.TYPE_OBJECT,
			// This won't work, because it has 'omitempty' on it.
			MaxProperties: 0,
		}, nil

	case "Int64Value", "Int32Value":
		return &jsonschema.Type{
			OneOf: []*jsonschema.Type{
				{ Type: gojsonschema.TYPE_NUMBER },
				{
					Type: gojsonschema.TYPE_STRING,
					Format: "regex",
					Pattern: `^-?[0-9]+$`,
				},
			},
		}, nil
		
	case "Timestamp":
		return &jsonschema.Type{
			Type: gojsonschema.TYPE_STRING,
			Format: "date-time",
		}, nil

	case "UInt64Value", "UInt32Value":
		return &jsonschema.Type{
			OneOf: []*jsonschema.Type{
				{ Type: gojsonschema.TYPE_NUMBER },
				{
					Type: gojsonschema.TYPE_STRING,
					Format: "regex",
					Pattern: `^[0-9]+$`,
				},
			},
		}, nil
	
	case "Value":
		return &jsonschema.Type{
			OneOf: []*jsonschema.Type{
				{Type: gojsonschema.TYPE_ARRAY},
				{Type: gojsonschema.TYPE_BOOLEAN},
				{Type: gojsonschema.TYPE_NUMBER},
				{Type: gojsonschema.TYPE_OBJECT},
				{Type: gojsonschema.TYPE_STRING},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown wkt: %s", name)
	}
}

// Converts a proto "MESSAGE" into a JSON-Schema:
func (c *Converter) convertMessageType(curPkg *ProtoPackage, msg *descriptorpb.DescriptorProto, pkgName string) (*jsonschema.Type, error) {
	if name := msg.GetName(); name != "" && pkgName == ".google.protobuf" {
		typ, err := convertWKT(name)
		if err != nil {
			return nil, err
		}

		switch {
		case typ.Type == gojsonschema.TYPE_NULL:
			return typ, nil

		case len(typ.OneOf) > 0:
			typ.OneOf = append(typ.OneOf, &jsonschema.Type{Type: gojsonschema.TYPE_NULL})
			return typ, nil
		}

		return &jsonschema.Type{
			OneOf: []*jsonschema.Type{
				typ,
				{Type: gojsonschema.TYPE_NULL},
			},
		}, nil
	}

	// Prepare a new jsonschema:
	typ := jsonschema.Type{
		Version: jsonschema.Version,
	}

	// Generate a description from src comments (if available)
	if src := c.sourceInfo.GetMessage(msg); src != nil {
		typ.Description = formatDescription(src)
	}

	// Optionally allow NULL values:
	if true {
		typ.OneOf = []*jsonschema.Type{
			{Type: gojsonschema.TYPE_OBJECT},
			{Type: gojsonschema.TYPE_NULL},
		}
	} else {
		typ.Type = gojsonschema.TYPE_OBJECT
	}

	// disallowAdditionalProperties will prevent validation where extra fields are found (outside of the schema):
	if c.DisallowAdditionalProperties {
		typ.AdditionalProperties = []byte("false")

	} else {
		typ.AdditionalProperties = []byte("true")
	}

	c.Logger.WithField("message_str", msg.String()).Trace("Converting message")

	for _, fieldDesc := range msg.GetField() {

		recursedTyp, err := c.convertField(curPkg, fieldDesc, msg)
		if err != nil {
			c.Logger.WithError(err).WithFields(logrus.Fields{
				"field_name": fieldDesc.GetName(),
				"message_name": msg.GetName(),
			}).Error("Failed to convert field")

			return nil, err
		}

		c.Logger.WithFields(logrus.Fields{
			"field_name": fieldDesc.GetName(),
			"type": recursedTyp.Type,
		}).Debug("Converted field")

		if typ.Properties == nil {
			typ.Properties = orderedmap.New()
		}

		typ.Properties.Set(fieldDesc.GetName(), recursedTyp)

		if c.UseProtoAndJSONFieldnames && fieldDesc.GetName() != fieldDesc.GetJsonName() {
			typ.Properties.Set(fieldDesc.GetJsonName(), recursedTyp)
		}
	}

	return &typ, nil
}

func formatDescription(sl *descriptorpb.SourceCodeInfo_Location) string {
	var lines []string
	for _, str := range sl.GetLeadingDetachedComments() {
		if s := strings.TrimSpace(str); s != "" {
			lines = append(lines, s)
		}
	}
	if s := strings.TrimSpace(sl.GetLeadingComments()); s != "" {
		lines = append(lines, s)
	}
	if s := strings.TrimSpace(sl.GetTrailingComments()); s != "" {
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n\n")
}
