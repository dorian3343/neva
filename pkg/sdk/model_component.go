/*
 * API Title
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package sdk

type Component struct {

	Io Io `json:"io,omitempty"`

	Workers map[string]string `json:"workers,omitempty"`

	Deps map[string]Io `json:"deps,omitempty"`

	Het []Connection `json:"het,omitempty"`
}

// AssertComponentRequired checks if the required fields are not zero-ed
func AssertComponentRequired(obj Component) error {
	if err := AssertIoRequired(obj.Io); err != nil {
		return err
	}
	for _, el := range obj.Het {
		if err := AssertConnectionRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseComponentRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of Component (e.g. [][]Component), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseComponentRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aComponent, ok := obj.(Component)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertComponentRequired(aComponent)
	})
}
