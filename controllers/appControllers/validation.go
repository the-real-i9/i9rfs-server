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
		validation.Field(&b.Command, validation.Required, validation.In(
			"list directory contents", "ls",
			"create new directory", "mkdir",
			"delete", "del",
			"trash",
			"restore",
			"show trash", "view trash",
			"rename",
			"move",
			"copy",
			"upload", "up",
		).Error("unrecognized command")),
		validation.Field(&b.CmdData, validation.Required.When(b.Command != "show trash" && b.Command != "view trash")),
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

type renameCmd struct {
	ParentDirectoryId string `json:"parentDirectoryId"`
	ObjectId          string `json:"objectId"`
	NewName           string `json:"newName"`
}

func (d renameCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectId, validation.Required, is.UUID),
		validation.Field(&d.NewName, validation.Required),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "renameCmd")
}

type moveCmd struct {
	FromParentDirectoryId string   `json:"fromParentDirectoryId"`
	ToParentDirectoryId   string   `json:"toParentDirectoryId"`
	ObjectIds             []string `json:"objectIds"`
}

func (d moveCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.FromParentDirectoryId, validation.Required, validation.When(d.FromParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ToParentDirectoryId, validation.Required, validation.When(d.ToParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "moveCmd")
}

type copyCmd struct {
	FromParentDirectoryId string   `json:"fromParentDirectoryId"`
	ToParentDirectoryId   string   `json:"toParentDirectoryId"`
	ObjectIds             []string `json:"objectIds"`
}

func (d copyCmd) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.FromParentDirectoryId, validation.Required, validation.When(d.FromParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ToParentDirectoryId, validation.Required, validation.When(d.ToParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "copyCmd")
}
