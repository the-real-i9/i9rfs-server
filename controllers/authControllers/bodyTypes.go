package authControllers

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

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
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b registerUserBody) Validate() error {

	return validation.ValidateStruct(&b,
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
