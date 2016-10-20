package model

import (
	"testing"
)

func TestString(t *testing.T) {
	name := "paus"
	registry := "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com"
	tag := "1.0.0"

	image := &Image{
		Name:     name,
		Registry: registry,
		Tag:      tag,
	}

	expected := "012345678901.dkr.ecr.ap-northeast-1.amazonaws.com/paus:1.0.0"
	actual := image.String()

	if actual != expected {
		t.Fatalf("Wrong image fullname. expected: %#v, actual: %#v", expected, actual)
	}
}
