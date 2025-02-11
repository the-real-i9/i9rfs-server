package appControllers

import (
	"i9rfs/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rfsCmdBody struct {
	Command string         `json:"command"`
	CmdData map[string]any `json:"data"`
}

func (b rfsCmdBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Command, validation.Required, validation.In().Error("unrecognized command")),
		validation.Field(&b.CmdData, validation.Required),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "rfsCmdBody")
}

type lsCmd struct {
	DirectoryId string `json:"directoryId"`
}

func (d lsCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.DirectoryId, validation.Required, validation.When(d.DirectoryId != "/", is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "lsCmd")
}

type mkdirCmd struct {
	ParentDirectoryId string `json:"parentDirectoryId"`
	DirectoryName     string `json:"directoryName"`
}

func (d mkdirCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.DirectoryName, validation.Required),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "mkdirCmd")
}

type delCmd struct {
	ParentDirectoryId string   `json:"parentDirectoryId"`
	ObjectIds         []string `json:"objectIds"`
}

func (d delCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "delCmd")
}

type trashCmd struct {
	ParentDirectoryId string   `json:"parentDirectoryId"`
	ObjectIds         []string `json:"objectIds"`
}

func (d trashCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "trashCmd")
}

type restoreCmd struct {
	ObjectIds []string `json:"objectIds"`
}

func (d restoreCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "restoreCmd")
}
