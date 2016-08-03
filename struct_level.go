package validator

import "reflect"

// StructLevelFunc accepts all values needed for struct level validation
type StructLevelFunc func(sl StructLevel)

// StructLevel contains all the information and helper functions
// to validate a struct
type StructLevel interface {

	// returns the main validation object, in case one want to call validations internally.
	// this is so you don;t have to use anonymous functoins to get access to the validate
	// instance.
	Validator() *Validate

	// returns the top level struct, if any
	Top() reflect.Value

	// returns the current fields parent struct, if any
	Parent() reflect.Value

	// returns the current struct.
	// this is not needed when implementing 'Validatable' interface,
	// only when a StructLevel is registered
	Current() reflect.Value

	// ExtractType gets the actual underlying type of field value.
	// It will dive into pointers, customTypes and return you the
	// underlying value and it's kind.
	ExtractType(field reflect.Value) (value reflect.Value, kind reflect.Kind, nullable bool)

	// reports an error just by passing the field and tag information
	//
	// NOTES:
	//
	// fieldName and altName get appended to the existing namespace that
	// validator is on. eg. pass 'FirstName' or 'Names[0]' depending
	// on the nesting
	//
	// tag can be an existing validation tag or just something you make up
	// and process on the flip side it's up to you.
	ReportError(field interface{}, fieldName, altName, tag string)

	// reports an error just by passing ValidationErrors
	//
	// NOTES:
	//
	// relativeNamespace and relativeActualNamespace get appended to the
	// existing namespace that validator is on.
	// eg. pass 'User.FirstName' or 'Users[0].FirstName' depending
	// on the nesting. most of the time they will be blank, unless you validate
	// at a level lower the the current field depth
	//
	// tag can be an existing validation tag or just something you make up
	// and process on the flip side it's up to you.
	ReportValidationErrors(relativeNamespace, relativeActualNamespace string, errs ValidationErrors)
}

var _ StructLevel = new(validate)

// Top returns the top level struct
//
// NOTE: this can be the same as the current struct being validated
// if not is a nested struct.
//
// this is only called when within Struct and Field Level validation and
// should not be relied upon for an acurate value otherwise.
func (v *validate) Top() reflect.Value {
	return v.top
}

// Parent returns the current structs parent
//
// NOTE: this can be the same as the current struct being validated
// if not is a nested struct.
//
// this is only called when within Struct and Field Level validation and
// should not be relied upon for an acurate value otherwise.
func (v *validate) Parent() reflect.Value {
	return v.slflParent
}

// Current returns the current struct.
func (v *validate) Current() reflect.Value {
	return v.slCurrent
}

// Validator returns the main validation object, in case one want to call validations internally.
func (v *validate) Validator() *Validate {
	return v.v
}

// ExtractType gets the actual underlying type of field value.
func (v *validate) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return v.extractTypeInternal(field, false)
}

// ReportError reports an error just by passing the field and tag information
func (v *validate) ReportError(field interface{}, fieldName, altName, tag string) {

	fv, kind, _ := v.extractTypeInternal(reflect.ValueOf(field), false)

	if len(altName) == 0 {
		altName = fieldName
	}

	ns := append(v.slNs, fieldName...)
	nsActual := append(v.slStructNs, altName...)

	switch kind {
	case reflect.Invalid:

		v.errs = append(v.errs,
			&fieldError{
				tag:         tag,
				actualTag:   tag,
				ns:          string(ns),
				structNs:    string(nsActual),
				field:       fieldName,
				structField: altName,
				param:       "",
				kind:        kind,
			},
		)

	default:

		v.errs = append(v.errs,
			&fieldError{
				tag:         tag,
				actualTag:   tag,
				ns:          string(ns),
				structNs:    string(nsActual),
				field:       fieldName,
				structField: altName,
				value:       fv.Interface(),
				param:       "",
				kind:        kind,
				typ:         fv.Type(),
			},
		)
	}
}

// ReportValidationErrors reports ValidationErrors obtained from running validations within the Struct Level validation.
//
// NOTE: this function prepends the current namespace to the relative ones.
func (v *validate) ReportValidationErrors(relativeNamespace, relativeActualNamespace string, errs ValidationErrors) {

	var err *fieldError

	for i := 0; i < len(errs); i++ {

		err = errs[i].(*fieldError)
		err.ns = string(append(append(v.slNs, err.ns...), err.field...))
		err.structNs = string(append(append(v.slStructNs, err.structNs...), err.structField...))

		v.errs = append(v.errs, err)
	}
}

// Validatable is the interface a struct can implement and
// be validated just like registering a StructLevel validation
// (they actually have the exact same signature.)
type Validatable interface {
	Validate(sl StructLevel)
}