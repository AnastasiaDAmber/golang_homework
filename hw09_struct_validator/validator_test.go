package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Numbers struct {
		Values []int `validate:"min:10|max:20"`
	}

	Bad struct {
		Field string `validate:"unknown:5"`
	}
)

func runValidationTest(t *testing.T, name string, input interface{}, expectedErr error) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		err := Validate(input)
		if !checkErrorContains(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestValidate_Users(t *testing.T) {
	runValidationTest(t, "valid user", User{
		ID:     "123456789012345678901234567890123456",
		Age:    25,
		Email:  "user@mail.com",
		Role:   "admin",
		Phones: []string{"89123456789", "89998887766"},
	}, nil)

	runValidationTest(t, "invalid id length", User{
		ID:     "short",
		Age:    25,
		Email:  "user@mail.com",
		Role:   "admin",
		Phones: []string{"89123456789"},
	}, ErrLen)

	runValidationTest(t, "invalid age too low", User{
		ID:     "123456789012345678901234567890123456",
		Age:    15,
		Email:  "user@mail.com",
		Role:   "admin",
		Phones: []string{"89123456789"},
	}, ErrMin)

	runValidationTest(t, "invalid age too high", User{
		ID:     "123456789012345678901234567890123456",
		Age:    99,
		Email:  "user@mail.com",
		Role:   "admin",
		Phones: []string{"89123456789"},
	}, ErrMax)

	runValidationTest(t, "invalid email", User{
		ID:     "123456789012345678901234567890123456",
		Age:    25,
		Email:  "bad_email",
		Role:   "admin",
		Phones: []string{"89123456789"},
	}, ErrRegexp)

	runValidationTest(t, "invalid role", User{
		ID:     "123456789012345678901234567890123456",
		Age:    25,
		Email:  "user@mail.com",
		Role:   "guest",
		Phones: []string{"89123456789"},
	}, ErrIn)

	runValidationTest(t, "invalid phone length", User{
		ID:     "123456789012345678901234567890123456",
		Age:    25,
		Email:  "user@mail.com",
		Role:   "admin",
		Phones: []string{"123"},
	}, ErrLen)
}

func TestValidate_App(t *testing.T) {
	runValidationTest(t, "valid app version", App{
		Version: "1.0.0",
	}, nil)

	runValidationTest(t, "invalid app version length", App{
		Version: "123",
	}, ErrLen)
}

func TestValidate_Response(t *testing.T) {
	runValidationTest(t, "response code not allowed", Response{
		Code: 201,
		Body: "ok",
	}, ErrIn)

	runValidationTest(t, "valid response code", Response{
		Code: 404,
		Body: "not found",
	}, nil)
}

func TestValidate_Other(t *testing.T) {
	runValidationTest(t, "not a struct (int)", 123, ErrNotAStruct)
	runValidationTest(t, "nil input", nil, ErrNotAStruct)
}

func TestValidate_MultipleErrors(t *testing.T) {
	runValidationTest(t, "multiple field errors", User{
		ID:     "short",
		Age:    5,
		Email:  "bademail",
		Role:   "guest",
		Phones: []string{"123"},
	}, ErrLen)
}

func TestValidate_IntSlice(t *testing.T) {
	runValidationTest(t, "valid int slice", Numbers{
		Values: []int{10, 15, 20},
	}, nil)

	runValidationTest(t, "invalid int slice element", Numbers{
		Values: []int{5, 25},
	}, ErrMin)
}

func TestValidate_Bad(t *testing.T) {
	runValidationTest(t, "unknown validator tag", Bad{
		Field: "test",
	}, fmt.Errorf("unknown validator unknown for string"))
}

// checkErrorContains как в твоем коде.
func checkErrorContains(err, expected error) bool {
	if err == nil {
		return expected == nil
	}
	if expected == nil {
		return false
	}

	var ve ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			if errors.Is(e.Err, expected) {
				return true
			}
			if e.Err.Error() == expected.Error() {
				return true
			}
		}
		return false
	}

	if errors.Is(err, expected) {
		return true
	}
	if err.Error() == expected.Error() {
		return true
	}
	return false
}
