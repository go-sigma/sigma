// Copyright 2023 XImager
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
