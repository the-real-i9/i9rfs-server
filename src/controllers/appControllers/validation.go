package appControllers

import (
	"i9rfs/src/helpers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type rfsActionBody struct {
	Action string         `json:"action"`
	Data   map[string]any `json:"data"`
}

func (b rfsActionBody) Validate() error {
	err := validation.ValidateStruct(&b,
		validation.Field(&b.Action, validation.Required),
		validation.Field(&b.Data, validation.Required.When(b.Action != "show trash" && b.Action != "view trash")),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "rfsActionBody")
}

type lsAction struct {
	DirectoryId string `json:"directoryId"`
}

func (d lsAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.DirectoryId, validation.Required, validation.When(d.DirectoryId != "/", is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "lsAction")
}

type mkdirAction struct {
	ParentDirectoryId string `json:"parentDirectoryId"`
	DirectoryName     string `json:"directoryName"`
}

func (d mkdirAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.DirectoryName, validation.Required),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "mkdirAction")
}

type delAction struct {
	ParentDirectoryId string   `json:"parentDirectoryId"`
	ObjectIds         []string `json:"objectIds"`
}

func (d delAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "delAction")
}

type trashAction struct {
	ParentDirectoryId string   `json:"parentDirectoryId"`
	ObjectIds         []string `json:"objectIds"`
}

func (d trashAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "trashAction")
}

type restoreAction struct {
	ObjectIds []string `json:"objectIds"`
}

func (d restoreAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "restoreAction")
}

type renameAction struct {
	ParentDirectoryId string `json:"parentDirectoryId"`
	ObjectId          string `json:"objectId"`
	NewName           string `json:"newName"`
}

func (d renameAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.ParentDirectoryId, validation.Required, validation.When(d.ParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectId, validation.Required, is.UUID),
		validation.Field(&d.NewName, validation.Required),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "renameAction")
}

type moveAction struct {
	FromParentDirectoryId string   `json:"fromParentDirectoryId"`
	ToParentDirectoryId   string   `json:"toParentDirectoryId"`
	ObjectIds             []string `json:"objectIds"`
}

func (d moveAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.FromParentDirectoryId, validation.Required, validation.When(d.FromParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ToParentDirectoryId, validation.Required, validation.When(d.ToParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "moveAction")
}

type copyAction struct {
	FromParentDirectoryId string   `json:"fromParentDirectoryId"`
	ToParentDirectoryId   string   `json:"toParentDirectoryId"`
	ObjectIds             []string `json:"objectIds"`
}

func (d copyAction) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.FromParentDirectoryId, validation.Required, validation.When(d.FromParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ToParentDirectoryId, validation.Required, validation.When(d.ToParentDirectoryId != "/", is.UUID)),
		validation.Field(&d.ObjectIds, validation.Required, validation.Each(is.UUID)),
	)

	return helpers.ValidationError(err, "appControllers_validation.go", "copyAction")
}
