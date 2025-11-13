package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	validationTag   = "validate"
	rulesSeparator  = "|"
	ruleSeparator   = ":"
	valuesSeparator = ","
)

var numericKinds = []reflect.Kind{
	reflect.Int,
}

var stringKinds = []reflect.Kind{
	reflect.String,
}

var arrayKinds = []reflect.Kind{
	reflect.Array,
	reflect.Slice,
}

var allowedFieldKinds = append(
	stringKinds,
	numericKinds...,
)

var (
	ErrArgumentNotStructure       = fmt.Errorf("argument is not a struct")
	ErrNonExistentValidationRule  = fmt.Errorf("non-existent validation rule")
	ErrNotSupportedValidationRule = fmt.Errorf("not supported validation rule")
)

var (
	ErrBaseValidation   = fmt.Errorf("validation error")
	ErrLenValidation    = fmt.Errorf("len %w", ErrBaseValidation)
	ErrMinValidation    = fmt.Errorf("min %w", ErrBaseValidation)
	ErrMaxValidation    = fmt.Errorf("max %w", ErrBaseValidation)
	ErrInListValidation = fmt.Errorf("in list %w", ErrBaseValidation)
	ErrRegexpValidation = fmt.Errorf("regexp %w", ErrBaseValidation)
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

	detailedErrs := make([]string, len(v))
	fields := make([]string, len(v))
	for i, validationError := range v {
		fields[i] = validationError.Field
		detailedErrs[i] = fmt.Sprintf("%v: %v", validationError.Field, validationError.Err)
	}
	return fmt.Sprintf(
		"invalid struct's fields with codes: %v.\n%v.",
		strings.Join(fields, ", "),
		strings.Join(detailedErrs, ".\n"),
	)
}

type Validator func(value reflect.Value, constraints []string) error

func Validate(v interface{}) error {
	refType := reflect.TypeOf(v)
	if refType.Kind() != reflect.Struct {
		return ErrArgumentNotStructure
	}

	validationErrs := ValidationErrors{}
	refValue := reflect.ValueOf(v)
	numFields := refType.NumField()
	for i := 0; i < numFields; i++ {
		field := refType.Field(i)
		if !isSupportedField(field) {
			continue
		}

		rulesUnparsed, ok := field.Tag.Lookup(validationTag)
		if !ok {
			continue
		}

		value := refValue.FieldByName(field.Name)
		var constraints []string
		ruleTags := strings.Split(rulesUnparsed, rulesSeparator)
		for _, ruleTag := range ruleTags {
			parsed := strings.Split(ruleTag, ruleSeparator)
			rule := parsed[0]
			validator, err := getValidatorByRule(rule, field)
			if err != nil {
				return err
			}

			if len(parsed) > 1 {
				constraints = strings.Split(parsed[1], valuesSeparator)
			} else {
				constraints = []string{}
			}

			// applying validator to field's value
			if !isKindInList(field.Type.Kind(), arrayKinds) {
				err = validator(value, constraints)
				if validationErrs, err = processErrorAfterValidation(err, validationErrs, field.Name); err != nil {
					return err
				}
				continue
			}

			for i := 0; i < value.Len(); i++ {
				err = validator(value.Index(i), constraints)
				fieldName := fmt.Sprintf("%v[%v]", field.Name, i)
				if validationErrs, err = processErrorAfterValidation(err, validationErrs, fieldName); err != nil {
					return err
				}
			}
		}
	}
	if len(validationErrs) == 0 {
		return nil
	}
	return validationErrs
}

// processErrorAfterValidation обработка ошибки после выполнения валидации значения поля.
func processErrorAfterValidation(
	err error,
	validationErrs ValidationErrors,
	fieldName string,
) (ValidationErrors, error) {
	if errors.Is(err, ErrBaseValidation) {
		validationErrs = append(
			validationErrs,
			ValidationError{
				Field: fieldName,
				Err:   err,
			},
		)
		return validationErrs, nil
	} else if err != nil {
		return validationErrs, err
	}
	return validationErrs, nil
}

// isSupportedField проверяет что тип поля поддерживается.
func isSupportedField(field reflect.StructField) bool {
	var kind reflect.Kind
	if isKindInList(field.Type.Kind(), arrayKinds) {
		kind = field.Type.Elem().Kind()
	} else {
		kind = field.Type.Kind()
	}
	return isKindInList(kind, allowedFieldKinds)
}

// isKindInList входит ли тип в список переданных.
func isKindInList(kind reflect.Kind, supported []reflect.Kind) bool {
	for _, allowed := range supported {
		if kind == allowed {
			return true
		}
	}
	return false
}

// getValidatorByRule возвращает Validator по коду правила и полю.
func getValidatorByRule(rule string, field reflect.StructField) (Validator, error) {
	getWrappedErr := func(err error) error {
		return fmt.Errorf(
			"get validator by rule '%v' for field '%v': %w",
			rule,
			field.Name,
			err,
		)
	}

	kind := field.Type.Kind()
	if isKindInList(kind, arrayKinds) {
		kind = field.Type.Elem().Kind()
	}

	switch rule {
	case "len":
		if !isKindInList(kind, stringKinds) {
			return nil, getWrappedErr(ErrNotSupportedValidationRule)
		}
		return lenCheck, nil
	case "regexp":
		if !isKindInList(kind, stringKinds) {
			return nil, getWrappedErr(ErrNotSupportedValidationRule)
		}
		return regexpCheck, nil
	case "in":
		return inCheck, nil
	case "min":
		if !isKindInList(kind, numericKinds) {
			return nil, getWrappedErr(ErrNotSupportedValidationRule)
		}
		return minCheck, nil
	case "max":
		if !isKindInList(kind, numericKinds) {
			return nil, getWrappedErr(ErrNotSupportedValidationRule)
		}
		return maxCheck, nil
	default:
		return nil, getWrappedErr(ErrNonExistentValidationRule)
	}
}

func lenCheck(value reflect.Value, constraints []string) error {
	lengthRaw := constraints[0]
	length, err := strconv.Atoi(lengthRaw)
	if err != nil {
		return fmt.Errorf("legth check: %w", err)
	}
	if len(value.String()) != length {
		return fmt.Errorf(
			"checks '%v' length, length '%v', required '%v': %w",
			value.String(),
			len(value.String()),
			length,
			ErrLenValidation,
		)
	}

	return nil
}

func minCheck(value reflect.Value, constraints []string) error {
	minRaw := constraints[0]
	minAcc, err := strconv.Atoi(minRaw)
	if err != nil {
		return fmt.Errorf("min check: %w", err)
	}
	if int(value.Int()) <= minAcc {
		return fmt.Errorf(
			"value '%v' below the minimum acceptable '%v': %w",
			value.Int(),
			minAcc,
			ErrMinValidation,
		)
	}

	return nil
}

func maxCheck(value reflect.Value, constraints []string) error {
	maxRaw := constraints[0]
	maxAcc, err := strconv.Atoi(maxRaw)
	if err != nil {
		return fmt.Errorf("max check: %w", err)
	}
	if int(value.Int()) >= maxAcc {
		return fmt.Errorf(
			"value '%v' above the maximum acceptable '%v': %w",
			value.Int(),
			maxAcc,
			ErrMaxValidation,
		)
	}

	return nil
}

func inCheck(value reflect.Value, constraints []string) error {
	for _, constraint := range constraints {
		switch value.Kind() { //nolint:exhaustive
		case reflect.Int:
			element, err := strconv.Atoi(constraint)
			if err != nil {
				return fmt.Errorf("in check: %w", err)
			}
			if element == int(value.Int()) {
				return nil
			}
			continue
		case reflect.String:
			if constraint == value.String() {
				return nil
			}
			continue
		default:
			return fmt.Errorf("in check: unsupported value kind")
		}
	}
	return fmt.Errorf(
		"value '%v' not in list of allowed values: '%v': %w",
		value,
		strings.Join(constraints, "', '"),
		ErrInListValidation,
	)
}

func regexpCheck(value reflect.Value, constraints []string) error {
	expr := constraints[0]
	regexpr, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	if regexpr.Match([]byte(value.String())) {
		return nil
	}

	return fmt.Errorf(
		"value '%v' mismatches with regexp '%v': %w",
		value,
		expr,
		ErrRegexpValidation,
	)
}
