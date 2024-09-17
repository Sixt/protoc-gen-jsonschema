package testdata

import _ "embed"

//go:embed ArrayOfEnums.jsonschema
var ArrayOfEnums string

//go:embed ArrayOfMessages.jsonschema
var ArrayOfMessages string

//go:embed ArrayOfObjects.jsonschema
var ArrayOfObjects string

//go:embed ArrayOfPrimitives.jsonschema
var ArrayOfPrimitives string

//go:embed Enumception.jsonschema
var Enumception string

//go:embed FirstEnum.jsonschema
var FirstEnum string

//go:embed FirstMessage.jsonschema
var FirstMessage string

//go:embed ImportedEnum.jsonschema
var ImportedEnum string

//go:embed Maps.jsonschema
var Maps string

//go:embed MessageWithComments.jsonschema
var MessageWithComments string

//go:embed NestedMessage.jsonschema
var NestedMessage string

//go:embed NestedObject.jsonschema
var NestedObject string

//go:embed PayloadMessage.jsonschema
var PayloadMessage string

//go:embed SecondEnum.jsonschema
var SecondEnum string

//go:embed SecondMessage.jsonschema
var SecondMessage string

//go:embed WellKnown.jsonschema
var WellKnown string

//go:embed ArrayOfPrimitivesDouble.jsonschema
var ArrayOfPrimitivesDouble string
