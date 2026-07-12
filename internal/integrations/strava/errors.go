package strava

import "errors"

var (
	errEmptyValue   = errors.New("empty value")
	errInvalidDate  = errors.New("invalid date")
	errInvalidBool  = errors.New("invalid boolean")
	errInvalidSport = errors.New("invalid sport type")
	errNoActivities = errors.New("activities.csv not found")
)
