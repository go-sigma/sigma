package dal

import "testing"

func TestDatabaseString(t *testing.T) {
	tests := []struct {
		name string
		x    Database
		want string
	}{
		{name: "postgresql", x: DatabasePostgresql, want: "postgresql"},
		{name: "mysql", x: DatabaseMysql, want: "mysql"},
		{name: "sqlite3", x: DatabaseSqlite3, want: "sqlite3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("Database.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseIsValid(t *testing.T) {
	tests := []struct {
		name string
		x    Database
		want bool
	}{
		{name: "postgresql", x: DatabasePostgresql, want: true},
		{name: "mysql", x: DatabaseMysql, want: true},
		{name: "sqlite3", x: DatabaseSqlite3, want: true},
		{name: "fake", x: Database("fake"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.IsValid(); got != tt.want {
				t.Errorf("Database.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDatabase(t *testing.T) {
	tests := []struct {
		name    string
		name2   string
		want    Database
		wantErr bool
	}{
		{name: "postgresql", name2: "postgresql", want: DatabasePostgresql, wantErr: false},
		{name: "mysql", name2: "mysql", want: DatabaseMysql, wantErr: false},
		{name: "sqlite3", name2: "sqlite3", want: DatabaseSqlite3, wantErr: false},
		{name: "fake", name2: "fake", want: Database(""), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDatabase(tt.name2)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDatabase() got = %v, want %v", got, tt.want)
			}
		})
	}
}
