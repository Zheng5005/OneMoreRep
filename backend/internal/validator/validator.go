package validator

import (
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// ValidateExercise validates exercise input fields.
func ValidateExercise(name, targetMuscle, notes string) error {
	name = strings.TrimSpace(name)
	targetMuscle = strings.TrimSpace(targetMuscle)
	notes = strings.TrimSpace(notes)

	errs := validation.Errors{}

	if err := validation.Validate(name, validation.Required, validation.Length(1, 255)); err != nil {
		errs["name"] = err
	}
	if err := validation.Validate(targetMuscle, validation.Required, validation.Length(1, 100)); err != nil {
		errs["target_muscle"] = err
	}
	if err := validation.Validate(notes, validation.Length(0, 2000)); err != nil {
		errs["notes"] = err
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// ValidateRoutineName validates a routine name.
func ValidateRoutineName(name string) error {
	name = strings.TrimSpace(name)
	return validation.Validate(name, validation.Required, validation.Length(1, 255))
}

// ValidateWorkoutSet validates workout set input fields.
func ValidateWorkoutSet(weight float64, reps int) error {
	errs := validation.Errors{}

	if err := validation.Validate(weight, validation.Required, validation.Min(0.0), validation.Max(9999.99)); err != nil {
		errs["weight"] = err
	}
	if err := validation.Validate(reps, validation.Required, validation.Min(1)); err != nil {
		errs["reps"] = err
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
