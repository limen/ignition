package ignition

import (
	"github.com/limen/ignition/validation"
	"reflect"
)

type Errors map[string][]string

type RequestEntity struct {
	Data     interface{}         // request data
	Rules    validation.Rules    // validation rules
	Requires validation.Requires // require fields
	Errors   Errors              // validation errors
}

// validate request data
func (e *RequestEntity) Validate() Errors {
	e.Errors = nil
	dataType := reflect.TypeOf(e.Data)
	dataValue := reflect.ValueOf(e.Data)
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		paramName := field.Tag.Get("json")
		if len(paramName) == 0 || !e.Rules.Has(paramName) {
			continue
		}
		v := dataValue.Field(i).Interface()
		if err := e.Rules[paramName].Match(v); err != nil {
			e.AddError(paramName, err.Error())
		}
	}

	return e.Errors
}

// add or append error
func (e *RequestEntity) AddError(field string, err string) {
	if e.Errors == nil {
		e.Errors = make(Errors)
	}

	if e.HaveFieldError(field) {
		e.Errors[field] = append(e.GetFieldError(field), err)
	} else {
		e.Errors[field] = []string{err}
	}
}

// Get field error
func (e *RequestEntity) GetFieldError(field string) []string {
	if e.Errors == nil {
		return nil
	}
	if err, ok := e.Errors[field]; ok {
		return err
	}

	return nil
}

// Check if field error exists
func (e *RequestEntity) HaveFieldError(f string) bool {
	if e.Errors == nil {
		return false
	}
	_, in := e.Errors[f]

	return in
}

// Check if the entity have errors
func (e *RequestEntity) HaveErrors() bool {
	return len(e.Errors) > 0
}
