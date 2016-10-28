package model

import (
	"testing"
)

func TestImageFromString(t *testing.T) {
	var (
		image *Image
		s     string
		err   error
	)

	var testdata = []struct {
		s        string
		registry string
		name     string
		tag      string
	}{
		{
			s:        "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com/paus",
			registry: "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com",
			name:     "paus",
			tag:      "",
		},
		{
			s:        "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com/paus:1.0.0",
			registry: "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com",
			name:     "paus",
			tag:      "1.0.0",
		},
		{
			s:        "dtan4/paus:1.0.0",
			registry: "",
			name:     "dtan4/paus",
			tag:      "1.0.0",
		},
		{
			s:        "quay.io/dtan4/paus:1.0.0",
			registry: "quay.io",
			name:     "dtan4/paus",
			tag:      "1.0.0",
		},
		{
			s:        "paus",
			registry: "",
			name:     "paus",
			tag:      "",
		},
		{
			s:        "paus:1.0.0",
			registry: "",
			name:     "paus",
			tag:      "1.0.0",
		},
	}

	for _, td := range testdata {
		image, err = ImageFromString(td.s)
		if err != nil {
			t.Fatalf("Error should not be raised. string: %#v, error: %#v", td.s, err)
		}

		if image.Registry != td.registry {
			t.Fatalf("Registry should be '"+td.registry+"'. actual: %#v", image.Registry)
		}

		if image.Name != td.name {
			t.Fatalf("Name should be '"+td.name+"'. actual: %#v", image.Name)
		}

		if image.Tag != td.tag {
			t.Fatalf("Tag should be '"+td.tag+"'. actual: %#v", image.Tag)
		}
	}

	s = "foobar:1.0.0:hoge"
	_, err = ImageFromString(s)
	if err == nil {
		t.Fatalf("Error should not raised. string: %#v", s)
	}
}

func TestString(t *testing.T) {
	registry := "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com"
	name := "paus"
	tag := "1.0.0"

	image := &Image{
		Registry: registry,
		Name:     name,
		Tag:      tag,
	}

	expected := "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com/paus:1.0.0"
	actual := image.String()

	if actual != expected {
		t.Fatalf("Wrong image fullname. expected: %#v, actual: %#v", expected, actual)
	}
}
