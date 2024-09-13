package converter

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestSourceInfoLookup(t *testing.T) {
	// Read in the test file & get references to the things we've declared.
	// Note that the hardcoded indexes must reflect the declaration order in
	// the .proto file.
	fds := mustReadProtoFiles(t, sampleProtoDirectory, "MessageWithComments.proto")

	protoFile := fds.File[0]
	msgWithComments := protoFile.MessageType[0]
	msgWithComments_name1 := msgWithComments.Field[0]

	// Create an instance of our thing and test that it returns the expected
	// source data for each of our above declarations.
	src := newSourceCodeInfo(fds.File)

	assertCommentsMatch(t, src.GetMessage(msgWithComments), &descriptorpb.SourceCodeInfo_Location{
		LeadingComments:         proto.String(" This is a message level comment and talks about what this message is and why you should care about it!\n"),
		LeadingDetachedComments: []string{" This is a leading detached comment (which becomes the title)\n"},
	})

	assertCommentsMatch(t, src.GetField(msgWithComments_name1), &descriptorpb.SourceCodeInfo_Location{
		LeadingComments: proto.String(" This field is supposed to represent blahblahblah\n"),
	})
}

func assertCommentsMatch(t *testing.T, actual, expected *descriptorpb.SourceCodeInfo_Location) {
	t.Helper()

	if len(actual.LeadingDetachedComments) != len(expected.LeadingDetachedComments) {
		t.Fatalf("Wrong length for LeadingDetachedComments.\n got: %v\nwant: %v", actual.LeadingDetachedComments, expected.LeadingDetachedComments)
	}

	for i := range actual.LeadingDetachedComments {
		if actual.LeadingDetachedComments[i] != expected.LeadingDetachedComments[i] {
			t.Fatalf("Wrong value for LeadingDetachedComments.\n got: %v\nwant: %v", actual.LeadingDetachedComments, expected.LeadingDetachedComments)
		}
	}

	if actual.GetTrailingComments() != expected.GetTrailingComments() {
		t.Fatalf("Wrong value for TrailingComments.\n got: %q\nwant: %q", actual.GetTrailingComments(), expected.GetTrailingComments())
	}

	if actual.GetLeadingComments() != expected.GetLeadingComments() {
		t.Fatalf("Wrong value for LeadingComments.\n got: %q\nwant: %q", actual.GetLeadingComments(), expected.GetLeadingComments())
	}
}
