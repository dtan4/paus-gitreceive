package model

import (
	"fmt"
	"strings"
)

const ecrDomainSuffix = "amazonaws.com"

type Image struct {
	Registry string
	Name     string
	Tag      string
}

func NewImage(registry, name, tag string) *Image {
	return &Image{
		Registry: registry,
		Name:     name,
		Tag:      tag,
	}
}

func ImageFromString(s string) (*Image, error) {
	var registry, name, tag string

	ss := strings.Split(s, ":")

	if len(ss) >= 3 {
		return nil, fmt.Errorf("Invalid image string. %s", s)
	}

	if len(ss) == 2 {
		tag = ss[1]
	}

	ss2 := strings.Split(ss[0], "/")

	if len(ss2) == 1 {
		name = ss2[0]
	} else if len(ss2) == 2 {
		if strings.HasSuffix(ss2[0], ecrDomainSuffix) {
			registry = ss2[0]
			name = ss2[1]
		} else {
			name = ss[0]
		}
	} else {
		registry = ss2[0]
		name = strings.SplitN(ss[0], "/", 2)[1]
	}

	return &Image{
		Registry: registry,
		Name:     name,
		Tag:      tag,
	}, nil
}

func (i *Image) String() string {
	return i.Registry + "/" + i.Name + ":" + i.Tag
}
