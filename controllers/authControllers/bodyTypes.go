package authControllers

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type signupBody struct {
	Step         string         `json:"step"`
	SessionToken string         `json:"sessionToken"`
	Data         map[string]any `json:"data"`
}

func (b signupBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Step,
			validation.Required,
			validation.In("one", "two", "three").Error("invalid step"),
		),
		validation.Field(&b.SessionToken,
			validation.When(b.Step != "one", validation.Required),
		),
		validation.Field(&b.Data,
			validation.Required,
		),
	)
}

type requestNewAccountBody struct {
	Email string `json:"email"`
}

func (b requestNewAccountBody) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.Email,
		),
	)
}

type verifyEmailBody struct {
	Code int `json:"code"`
}

func (b verifyEmailBody) Validate() error {
	mb := struct {
		Code string `json:"code"`
	}{Code: fmt.Sprint(b.Code)}

	return validation.ValidateStruct(&mb,
		validation.Field(&mb.Code,
			validation.Required,
			validation.Length(6, 6).Error("invalid code value"),
		),
	)
}

type registerUserBody struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b registerUserBody) Validate() error {

	return validation.ValidateStruct(&b,
		validation.Field(&b.Email,
			validation.Required,
			is.Email.Error("invalid email format"),
		),
		validation.Field(&b.Username,
			validation.Required,
			validation.Length(3, 0).Error("username too short"),
			validation.Match(regexp.MustCompile("^[[:alnum:]][[:alnum:]_-]+[[:alnum:]]$")).Error("invalid username syntax"),
		),
		validation.Field(&b.Password,
			validation.Required,
			validation.Length(8, 0).Error("minimum of 8 characters"),
		),
	)
}

type signInBody struct {
	EmailOrUsername string `json:"emailOrUsername"`
	Password        string `json:"password"`
}

func (b signInBody) Validate() error {

	return validation.ValidateStruct(&b,
		validation.Field(&b.EmailOrUsername,
			validation.Required,
			validation.When(strings.ContainsAny(b.EmailOrUsername, "@"),
				is.Email.Error("invalid email or username"),
			).Else(
				validation.Length(3, 0).Error("invalid email or username"),
			),
		),
		validation.Field(&b.Password,
			validation.Required,
		),
	)
}
