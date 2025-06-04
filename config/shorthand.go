package config

import (
	"encoding/json"
	"errors"
)

// unmarshalShorthand unmarshals JSON representation of a given slice type
// that was either provided as a JSON array of strings or an array of objects.
func unmarshalShorthand[S ~[]E, E any](unmarshal []byte, factory func(string) E) (S, error) {
	shorthand := []string{}
	err := json.Unmarshal(unmarshal, &shorthand)
	s := new(S)

	if err == nil {
		// Input was entirely in short form
		*s = make(S, len(shorthand))

		for i, value := range shorthand {
			(*s)[i] = factory(value)
		}

		return *s, nil
	}

	// We treat UnmarshalTypeError as a soft error. It means that some part of
	// the input could not be matched to the target interface. Other errors
	// indicate severly malformed input, so we will propagate the error.
	if !IsUnmarshalTypeError(err) {
		return *s, err
	}

	longhand := []E{}
	err = json.Unmarshal(unmarshal, &longhand)

	if err == nil {
		// Input was entirely in long form
		return S(longhand), nil
	}

	if !IsUnmarshalTypeError(err) {
		return *s, err
	}

	if len(shorthand) != len(longhand) {
		return *s, errors.New("mismatched unmarshal results")
	}

	// Input was mixed short and long form. Walk the short form results and
	// turn any non-empty strings into E values in the same slot
	// of the long form results.
	for i, value := range shorthand {
		if value != "" {
			longhand[i] = factory(value)
		}
	}

	return S(longhand), nil
}
