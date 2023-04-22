// Code generated by go-enum DO NOT EDIT.
// Version: 0.5.6
// Revision: 97611fddaa414f53713597918c5e954646cb8623
// Build Date: 2023-03-26T21:38:06Z
// Built By: goreleaser

package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

const (
	// TaskCommonStatusPending is a TaskCommonStatus of type Pending.
	TaskCommonStatusPending TaskCommonStatus = "Pending"
	// TaskCommonStatusDoing is a TaskCommonStatus of type Doing.
	TaskCommonStatusDoing TaskCommonStatus = "Doing"
	// TaskCommonStatusSuccess is a TaskCommonStatus of type Success.
	TaskCommonStatusSuccess TaskCommonStatus = "Success"
	// TaskCommonStatusFailed is a TaskCommonStatus of type Failed.
	TaskCommonStatusFailed TaskCommonStatus = "Failed"
)

var ErrInvalidTaskCommonStatus = errors.New("not a valid TaskCommonStatus")

// String implements the Stringer interface.
func (x TaskCommonStatus) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x TaskCommonStatus) IsValid() bool {
	_, err := ParseTaskCommonStatus(string(x))
	return err == nil
}

var _TaskCommonStatusValue = map[string]TaskCommonStatus{
	"Pending": TaskCommonStatusPending,
	"Doing":   TaskCommonStatusDoing,
	"Success": TaskCommonStatusSuccess,
	"Failed":  TaskCommonStatusFailed,
}

// ParseTaskCommonStatus attempts to convert a string to a TaskCommonStatus.
func ParseTaskCommonStatus(name string) (TaskCommonStatus, error) {
	if x, ok := _TaskCommonStatusValue[name]; ok {
		return x, nil
	}
	return TaskCommonStatus(""), fmt.Errorf("%s is %w", name, ErrInvalidTaskCommonStatus)
}

// MustParseTaskCommonStatus converts a string to a TaskCommonStatus, and panics if is not valid.
func MustParseTaskCommonStatus(name string) TaskCommonStatus {
	val, err := ParseTaskCommonStatus(name)
	if err != nil {
		panic(err)
	}
	return val
}

var errTaskCommonStatusNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *TaskCommonStatus) Scan(value interface{}) (err error) {
	if value == nil {
		*x = TaskCommonStatus("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseTaskCommonStatus(v)
	case []byte:
		*x, err = ParseTaskCommonStatus(string(v))
	case TaskCommonStatus:
		*x = v
	case *TaskCommonStatus:
		if v == nil {
			return errTaskCommonStatusNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errTaskCommonStatusNilPtr
		}
		*x, err = ParseTaskCommonStatus(*v)
	default:
		return errors.New("invalid type for TaskCommonStatus")
	}

	return
}

// Value implements the driver Valuer interface.
func (x TaskCommonStatus) Value() (driver.Value, error) {
	return x.String(), nil
}
