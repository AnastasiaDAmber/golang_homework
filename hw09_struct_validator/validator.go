package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrLen        = errors.New("len validation failed")
	ErrRegexp     = errors.New("regexp validation failed")
	ErrIn         = errors.New("in validation failed")
	ErrMin        = errors.New("min validation failed")
	ErrMax        = errors.New("max validation failed")
	ErrNotAStruct = errors.New("input must be struct or pointer to struct")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, e := range v {
		sb.WriteString(fmt.Sprintf("%s: %v", e.Field, e.Err))
		if i != len(v)-1 {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

// ---------------- Универсальные валидаторы ----------------

// ValidateValue — универсальный валидатор для одного значения.
func ValidateValue[T any](v T, rules string) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return validateString(rv.String(), rules)
	case reflect.Int:
		return validateInt(int(rv.Int()), rules)
	case reflect.Invalid:
		return fmt.Errorf("invalid value")
	case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Ptr, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
		return fmt.Errorf("unsupported type %T", v)
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
}

func validateString(value string, rules string) error {
	validators := strings.Split(rules, "|")
	for _, rule := range validators {
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid rule: %s", rule)
		}
		name, param := parts[0], parts[1]

		switch name {
		case "len":
			expected, _ := strconv.Atoi(param)
			if len(value) != expected {
				return fmt.Errorf("%w: expected %d, got %d", ErrLen, expected, len(value))
			}
		case "regexp":
			re := regexp.MustCompile(param)
			if !re.MatchString(value) {
				return fmt.Errorf("%w: %s does not match %s", ErrRegexp, value, param)
			}
		case "in":
			opts := strings.Split(param, ",")
			found := false
			for _, opt := range opts {
				if value == opt {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%w: %s not in [%s]", ErrIn, value, param)
			}
		default:
			// тут ключевой момент: возвращаем ошибку напрямую
			return fmt.Errorf("unknown validator %s for string", name)
		}
	}
	return nil
}

func validateInt(value int, rules string) error {
	validators := strings.Split(rules, "|")
	for _, rule := range validators {
		if rule == "" {
			continue
		}
		parts := strings.SplitN(rule, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid rule: %s", rule)
		}
		name, param := parts[0], parts[1]

		switch name {
		case "min":
			minVal, _ := strconv.Atoi(param)
			if value < minVal {
				return fmt.Errorf("%w: %d < %d", ErrMin, value, minVal)
			}
		case "max":
			maxVal, _ := strconv.Atoi(param)
			if value > maxVal {
				return fmt.Errorf("%w: %d > %d", ErrMax, value, maxVal)
			}
		case "in":
			opts := strings.Split(param, ",")
			found := false
			for _, opt := range opts {
				n, _ := strconv.Atoi(strings.TrimSpace(opt))
				if value == n {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%w: %d not in [%s]", ErrIn, value, param)
			}
		default:
			return fmt.Errorf("unknown validator %s for int", name)
		}
	}
	return nil
}

// ValidateSlice — универсальный валидатор для слайсов.
func ValidateSlice[T any](values []T, rules string) error {
	var errs ValidationErrors
	for i, v := range values {
		if err := ValidateValue(v, rules); err != nil {
			// добавляем каждую ошибку в errs
			var ve ValidationErrors
			if errors.As(err, &ve) {
				for _, e := range ve {
					errs = append(errs, ValidationError{
						Field: fmt.Sprintf("elem[%d].%s", i, e.Field),
						Err:   e.Err,
					})
				}
			} else {
				errs = append(errs, ValidationError{
					Field: fmt.Sprintf("elem[%d]", i),
					Err:   err,
				})
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateStruct — основной валидатор для структур.
func ValidateStruct[T any](s T) error {
	val := reflect.ValueOf(s)
	if !val.IsValid() {
		return ErrNotAStruct
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ErrNotAStruct
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	typ := val.Type()
	var errs ValidationErrors

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		var err error

		switch fieldValue.Kind() {
		case reflect.String:
			err = ValidateValue(fieldValue.String(), tag)
		case reflect.Int:
			err = ValidateValue(int(fieldValue.Int()), tag)
		case reflect.Slice:
			//nolint:exhaustive
			switch fieldValue.Type().Elem().Kind() {
			case reflect.String:
				err = ValidateSlice(fieldValue.Interface().([]string), tag)
			case reflect.Int:
				err = ValidateSlice(fieldValue.Interface().([]int), tag)
			default:
				err = fmt.Errorf("unsupported slice element type: %s", fieldValue.Type().Elem().Kind())
			}
		case reflect.Invalid:
			err = fmt.Errorf("invalid field value")
		case reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
			reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Struct, reflect.UnsafePointer:
			err = fmt.Errorf("unsupported type %s", fieldValue.Kind())
		default:
			err = fmt.Errorf("unsupported type %s", fieldValue.Kind())
		}

		if err != nil {
			var ve ValidationErrors
			if errors.As(err, &ve) {
				for _, e := range ve {
					errs = append(errs, ValidationError{
						Field: fmt.Sprintf("%s.%s", field.Name, e.Field),
						Err:   e.Err,
					})
				}
			} else {
				errs = append(errs, ValidationError{
					Field: field.Name,
					Err:   err,
				})
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Validate — универсальная точка входа.
func Validate[T any](v T) error {
	return ValidateStruct(v)
}
