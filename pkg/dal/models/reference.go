package models

// Reference is the model for reference.
type Reference struct {
	Artifact *Artifact
	Tag      []*Tag
}
