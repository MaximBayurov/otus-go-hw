package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	UnsupportedTags struct {
		Title string `validate:"inList:200,404,500"`
		Text  string `validate:"omitempty"`
	}

	PrivateFields struct {
		title string `validate:"len:10"`
		text  string `validate:"len:10"`
	}
)

func TestValidate(t *testing.T) {
	tests := getTestCases()

	for i, tt := range tests {
		var name string
		if len(tt.name) > 0 {
			name = tt.name
		} else {
			name = fmt.Sprintf("case %d", i)
		}
		t.Run(name, func(t *testing.T) {
			tt := tt
			t.Parallel()
			err := Validate(tt.in)
			if tt.expectedErr == nil {
				assert.NoError(t, err, "unexpected error")
				return
			}

			assertErrs := ValidationErrors{}
			validationErr := ValidationErrors{}
			if errors.As(err, &validationErr) && errors.As(tt.expectedErr, &assertErrs) {
				for i, validationError := range validationErr {
					assertErr := assertErrs[i]

					assert.True(t, errors.Is(validationError.Err, assertErr.Err))
					assert.Equal(t, assertErr.Field, validationError.Field)
				}
				return
			}

			assert.True(t, errors.Is(err, tt.expectedErr))

			_ = tt
		})
	}
}

type testCase struct {
	in          interface{}
	expectedErr error
	name        string
}

func getTestCases() []testCase {
	return []testCase{
		{
			in: User{
				ID:    "123",
				Name:  "Test1",
				Age:   15,
				Email: "valid@mail.ru",
				Role:  "user",
				Phones: []string{
					"12345678901",
					"1234567890",
					"123456789",
				},
				meta: nil,
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "ID",
					Err:   ErrLenValidation,
				},
				ValidationError{
					Field: "Age",
					Err:   ErrMinValidation,
				},
				ValidationError{
					Field: "Role",
					Err:   ErrInListValidation,
				},
				ValidationError{
					Field: "Phones[1]",
					Err:   ErrLenValidation,
				},
				ValidationError{
					Field: "Phones[2]",
					Err:   ErrLenValidation,
				},
			},
		},
		{
			in: User{
				ID:    "000000000000000000000000000000000123",
				Name:  "Test2",
				Age:   51,
				Email: "invalidmailru",
				Role:  "admin",
				Phones: []string{
					"12345678901",
					"09876543211",
				},
				meta: nil,
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Age",
					Err:   ErrMaxValidation,
				},
				ValidationError{
					Field: "Email",
					Err:   ErrRegexpValidation,
				},
			},
		},
		{in: App{Version: "1.1.1"}, expectedErr: nil},
		{in: App{Version: "1.1.12"}, expectedErr: ValidationErrors{
			ValidationError{
				Field: "Version",
				Err:   ErrLenValidation,
			},
		}},
		{
			in: Token{
				Header:    nil,
				Payload:   nil,
				Signature: nil,
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 200,
				Body: "",
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 301,
				Body: "",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Code",
					Err:   ErrInListValidation,
				},
			},
		},
		{
			in: UnsupportedTags{
				Title: "title",
				Text:  "text",
			},
			expectedErr: ErrNonExistentValidationRule,
		},
		{
			in: PrivateFields{
				title: "132",
				text:  "",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "title",
					Err:   ErrLenValidation,
				},
				ValidationError{
					Field: "text",
					Err:   ErrLenValidation,
				},
			},
		},
	}
}
