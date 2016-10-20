package model

type Image struct {
	Name     string
	Registry string
	Tag      string
}

func NewImage(name, registry, tag string) *Image {
	return &Image{
		Name:     name,
		Registry: registry,
		Tag:      tag,
	}
}

func (i *Image) String() string {
	return i.Registry + "/" + i.Name + ":" + i.Tag
}
