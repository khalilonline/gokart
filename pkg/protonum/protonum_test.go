package protonum

import (
	"testing"

	"github.com/khalilonline/gokart/pkg/testflags"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/descriptorpb"
)

// labelName stands in for a domain package's own name type.
type labelName string

type label = descriptorpb.FieldDescriptorProto_Label

const (
	optional = descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	required = descriptorpb.FieldDescriptorProto_LABEL_REQUIRED
	repeated = descriptorpb.FieldDescriptorProto_LABEL_REPEATED
)

func labels() map[label]labelName {
	return map[label]labelName{optional: "optional", required: "required", repeated: "repeated"}
}

type protonumTestSuite struct {
	suite.Suite
}

func TestProtonumSuite(t *testing.T) {
	testflags.UnitTest(t)
	suite.Run(t, new(protonumTestSuite))
}

func (s *protonumTestSuite) TestNameAndEnum() {
	tr := New(labels())
	for e, name := range labels() {
		got, ok := tr.Name(e)
		s.True(ok, "Name(%v)", e)
		s.Equal(name, got, "Name(%v)", e)

		back, ok := tr.Enum(name)
		s.True(ok, "Enum(%q)", name)
		s.Equal(e, back, "Enum(%q)", name)
	}
}

func (s *protonumTestSuite) TestUnnamed() {
	tr := New(labels())
	for _, e := range []label{0, 9} {
		name, ok := tr.Name(e)
		s.False(ok, "Name(%v) = %q; want no name", e, name)
	}
	e, ok := tr.Enum("packed")
	s.False(ok, `Enum("packed") = %v; want no value`, e)
}

func (s *protonumTestSuite) TestNameOf() {
	tr := New(labels())
	s.Equal(labelName("required"), tr.NameOf(required))
	s.Equal(labelName(""), tr.NameOf(0))
}

func (s *protonumTestSuite) TestAllInNumberOrder() {
	s.Equal([]labelName{"optional", "required", "repeated"}, New(labels()).All())
}

func (s *protonumTestSuite) TestNames() {
	tr := New(labels())
	names, err := tr.Names([]label{repeated, optional})
	s.Require().NoError(err)
	s.Equal([]labelName{"repeated", "optional"}, names)

	names, err = tr.Names(nil)
	s.Require().NoError(err)
	s.Nil(names)

	_, err = tr.Names([]label{optional, 0})
	s.ErrorIs(err, ErrUnnamed)
}

func (s *protonumTestSuite) TestKnownNames() {
	tr := New(labels())
	s.Equal([]labelName{"optional", "repeated"}, tr.KnownNames([]label{optional, 9, repeated}))
	s.Nil(tr.KnownNames(nil))
}

func (s *protonumTestSuite) TestEnums() {
	tr := New(labels())
	es, err := tr.Enums([]labelName{"required", "optional"})
	s.Require().NoError(err)
	s.Equal([]label{required, optional}, es)

	es, err = tr.Enums(nil)
	s.Require().NoError(err)
	s.Nil(es)

	_, err = tr.Enums([]labelName{"optional", "packed"})
	s.ErrorIs(err, ErrUnknownName)
}

func (s *protonumTestSuite) TestCopiesTheNames() {
	names := labels()
	tr := New(names)
	names[optional] = "changed"

	got, _ := tr.Name(optional)
	s.Equal(labelName("optional"), got)
}

func (s *protonumTestSuite) TestNewPanics() {
	tests := []struct {
		name  string
		build func()
		want  string
	}{
		{
			name: "a value without a name",
			build: func() {
				names := labels()
				delete(names, required)
				New(names)
			},
			want: "protonum: google.protobuf.FieldDescriptorProto.LABEL_REQUIRED has no name",
		},
		{
			name: "a name for the zero value",
			build: func() {
				New(map[descriptorpb.FieldOptions_CType]string{
					descriptorpb.FieldOptions_STRING:       "string",
					descriptorpb.FieldOptions_CORD:         "cord",
					descriptorpb.FieldOptions_STRING_PIECE: "string_piece",
				})
			},
			want: "protonum: google.protobuf.FieldOptions.STRING gets a name",
		},
		{
			name: "a number the enum does not define",
			build: func() {
				names := labels()
				names[9] = "nine"
				New(names)
			},
			want: "protonum: google.protobuf.FieldDescriptorProto.Label has no value 9",
		},
		{
			name: "the zero number of an enum that does not define it",
			build: func() {
				names := labels()
				names[0] = "unset"
				New(names)
			},
			want: "protonum: google.protobuf.FieldDescriptorProto.Label has no value 0",
		},
		{
			name: "a name two values share",
			build: func() {
				names := labels()
				names[repeated] = "optional"
				New(names)
			},
			want: `protonum: google.protobuf.FieldDescriptorProto.LABEL_OPTIONAL and google.protobuf.FieldDescriptorProto.LABEL_REPEATED share the name "optional"`,
		},
		{
			name: "an empty name",
			build: func() {
				names := labels()
				names[repeated] = ""
				New(names)
			},
			want: "protonum: google.protobuf.FieldDescriptorProto.LABEL_REPEATED has an empty name",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.PanicsWithValue(tt.want, tt.build)
		})
	}
}
