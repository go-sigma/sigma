package types

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name string
		x    TaskCommonStatus
		want string
	}{
		{name: "Pending", x: TaskCommonStatusPending, want: "Pending"},
		{name: "Doing", x: TaskCommonStatusDoing, want: "Doing"},
		{name: "Success", x: TaskCommonStatusSuccess, want: "Success"},
		{name: "Failed", x: TaskCommonStatusFailed, want: "Failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Database.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		x    TaskCommonStatus
		want bool
	}{
		{name: "Pending", x: TaskCommonStatusPending, want: true},
		{name: "Doing", x: TaskCommonStatusDoing, want: true},
		{name: "Success", x: TaskCommonStatusSuccess, want: true},
		{name: "Failed", x: TaskCommonStatusFailed, want: true},
		{name: "Invalid", x: TaskCommonStatus("Invalid"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.IsValid(); got != tt.want {
				t.Errorf("Database.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTaskCommonStatus(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    TaskCommonStatus
		wantErr bool
	}{
		{name: "Pending", arg: "Pending", want: TaskCommonStatusPending},
		{name: "Doing", arg: "Doing", want: TaskCommonStatusDoing},
		{name: "Success", arg: "Success", want: TaskCommonStatusSuccess},
		{name: "Failed", arg: "Failed", want: TaskCommonStatusFailed},
		{name: "Invalid", arg: "Invalid", want: TaskCommonStatus(""), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskCommonStatus(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskCommonStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTaskCommonStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustParseTaskCommonStatus(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want TaskCommonStatus
	}{
		{name: "Pending", arg: "Pending", want: TaskCommonStatusPending},
		{name: "Doing", arg: "Doing", want: TaskCommonStatusDoing},
		{name: "Success", arg: "Success", want: TaskCommonStatusSuccess},
		{name: "Failed", arg: "Failed", want: TaskCommonStatusFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustParseTaskCommonStatus(tt.arg); got != tt.want {
				t.Errorf("MustParseTaskCommonStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustParseTaskCommonStatusPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParseTaskCommonStatus() did not panic")
		}
	}()
	MustParseTaskCommonStatus("Invalid")
}
