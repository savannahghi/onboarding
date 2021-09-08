package domain

import (
	"fmt"
	"io"
	"log"
	"strconv"
)

// FivePointRating is used to implement
type FivePointRating string

// known ratings
const (
	FivePointRatingPoor           FivePointRating = "POOR"
	FivePointRatingUnsatisfactory FivePointRating = "UNSATISFACTORY"
	FivePointRatingAverage        FivePointRating = "AVERAGE"
	FivePointRatingSatisfactory   FivePointRating = "SATISFACTORY"
	FivePointRatingExcellent      FivePointRating = "EXCELLENT"
)

// AllFivePointRating is a list of all known ratings
var AllFivePointRating = []FivePointRating{
	FivePointRatingPoor,
	FivePointRatingUnsatisfactory,
	FivePointRatingAverage,
	FivePointRatingSatisfactory,
	FivePointRatingExcellent,
}

// IsValid returns true for valid ratings
func (e FivePointRating) IsValid() bool {
	switch e {
	case FivePointRatingPoor, FivePointRatingUnsatisfactory, FivePointRatingAverage, FivePointRatingSatisfactory, FivePointRatingExcellent:
		return true
	}
	return false
}

func (e FivePointRating) String() string {
	return string(e)
}

// UnmarshalGQL converts the input, if valid, into a rating value
func (e *FivePointRating) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FivePointRating(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FivePointRating", str)
	}
	return nil
}

// MarshalGQL converts the rating into a valid JSON string
func (e FivePointRating) MarshalGQL(w io.Writer) {
	_, err := fmt.Fprint(w, strconv.Quote(e.String()))
	if err != nil {
		log.Printf("%v\n", err)
	}
}

// EmploymentType ...
type EmploymentType string

// EmploymentTypeEmployed ..
const (
	EmploymentTypeEmployed     EmploymentType = "EMPLOYED"
	EmploymentTypeSelfEmployed EmploymentType = "SELF_EMPLOYED"
)

// AllEmploymentType ..
var AllEmploymentType = []EmploymentType{
	EmploymentTypeEmployed,
	EmploymentTypeSelfEmployed,
}

// IsValid ..
func (e EmploymentType) IsValid() bool {
	switch e {
	case EmploymentTypeEmployed, EmploymentTypeSelfEmployed:
		return true
	}
	return false
}

func (e EmploymentType) String() string {
	return string(e)
}

// UnmarshalGQL ..
func (e *EmploymentType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = EmploymentType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid EmploymentType", str)
	}
	return nil
}

// MarshalGQL ..
func (e EmploymentType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
