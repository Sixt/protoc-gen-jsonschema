// protoc plugin which converts .proto to JSON schema
// It is spawned by protoc and generates JSON-schema files.
// "Heavily influenced" by Google's "protog-gen-bq-schema"
//
// usage:
//  $ bin/protoc --jsonschema_out=path/to/outdir foo.proto
//
package main

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"github.com/sirupsen/logrus"
	"github.com/sixt/protoc-gen-jsonschema/internal/converter"
)

func main() {

	// Make a Logrus logger (default to INFO):
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(os.Stderr)

	// Use the logger to make a Converter:
	protoConverter := &converter.Converter{
		Logger: logger,
	}

	// Convert the generator request:
	var failed bool
	logger.Debug("Processing code generator request")
	res, err := protoConverter.ConvertFrom(os.Stdin)
	if err != nil {
		failed = true
		if res == nil {
			res = &pluginpb.CodeGeneratorResponse{
				Error: proto.String(fmt.Sprint("Failed to read input:", err)),
			}
		}
	}

	logger.Debug("Serializing code generator response")
	data, err := proto.Marshal(res)
	if err != nil {
		logger.WithError(err).Fatal("Cannot marshal response")
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		logger.WithError(err).Fatal("Failed to write response")
	}

	if failed {
		logger.Warn("Failed to process code generator but successfully sent the error to protoc")
		os.Exit(1)
	}

	logger.Debug("Succeeded to process code generator request")
}
