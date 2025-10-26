package sql_adapter

import (
	"database/sql"
	"fmt"
	"runtime"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// CasbinRule ...
type CasbinRule struct {
	ID    int64  `db:"id"`
	PType string `db:"ptype"`
	V0    string `db:"v0"`
	V1    string `db:"v1"`
	V2    string `db:"v2"`
	V3    string `db:"v3"`
	V4    string `db:"v4"`
	V5    string `db:"v5"`
}

// Adapter represents the sqlx adapter for policy storage.
type Adapter struct {
	db         *sql.DB
	tableName  string
	isFiltered bool
}

// Filter ...
type Filter struct {
	PType []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

// // AdapterOptions contains all possible configuration options.
// type AdapterOptions {
// 	DriverName     string
// 	DataSourceName string
// 	TableName      string
// 	DB             *sqlx.DB
// }

func finalizer(a *Adapter) {
	a.db.Close()
}

func loadPolicyLine(line CasbinRule, model model.Model) {
	lineText := line.PType
	if line.V0 != "" {
		lineText += ", " + line.V0
	}
	if line.V1 != "" {
		lineText += ", " + line.V1
	}
	if line.V2 != "" {
		lineText += ", " + line.V2
	}
	if line.V3 != "" {
		lineText += ", " + line.V3
	}
	if line.V4 != "" {
		lineText += ", " + line.V4
	}
	if line.V5 != "" {
		lineText += ", " + line.V5
	}
	persist.LoadPolicyLine(lineText, model) // nolint: errcheck
}

func savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{}
	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}
	return line
}

func (a *Adapter) dropTable() {
	_, err := a.db.Exec(fmt.Sprintf("DELETE FROM %s", a.tableName))
	if err != nil {
		panic(err)
	}
}

func (a *Adapter) ensureTable() {
	_, err := a.db.Exec(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", a.tableName))
	if err != nil {
		panic(err)
	}
}

func (a *Adapter) insertPolicyLine(line *CasbinRule) (err error) {
	query := fmt.Sprintf("INSERT INTO %s (ptype, v0, v1, v2, v3, v4, v5) VALUES (?, ?, ?, ?, ?, ?, ?)", a.tableName) // nolint: gosec
	_, err = a.db.Exec(query, line.PType, line.V0, line.V1, line.V2, line.V3, line.V4, line.V5)
	if err != nil {
		return
	}
	return
}

func (a *Adapter) deletePolicyLine(line *CasbinRule) (err error) {
	query := fmt.Sprintf( // nolint: gosec
		"DELETE FROM %s WHERE ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
		a.tableName,
	)
	_, err = a.db.Exec(query, line.PType, line.V0, line.V1, line.V2, line.V3, line.V4, line.V5)
	if err != nil {
		return
	}
	return
}

// NewAdapterByDB is the constructor for Adapter with existed connection
// Deprecated: Use NewAdapterFromOptions instead
func NewAdapterByDB(db *sql.DB) *Adapter {
	a := &Adapter{
		db:        db,
		tableName: "casbin_rules",
	}
	a.ensureTable()
	runtime.SetFinalizer(a, finalizer)
	return a
}

// LoadPolicy loads policy from database.
func (a *Adapter) LoadPolicy(model model.Model) error {
	var lines []CasbinRule
	rows, err := a.db.Query(fmt.Sprintf("SELECT id, ptype, v0, v1, v2, v3, v4, v5 FROM %s", a.tableName))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var line CasbinRule
		err := rows.Scan(&line.ID, &line.PType, &line.V0, &line.V1, &line.V2, &line.V3, &line.V4, &line.V5)
		if err != nil {
			return err
		}
		lines = append(lines, line)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	for _, line := range lines {
		loadPolicyLine(line, model)
	}
	return nil
}

// SavePolicy saves policy to database.
func (a *Adapter) SavePolicy(model model.Model) (err error) {
	a.dropTable()
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			err = a.insertPolicyLine(&line)
			if err != nil {
				return
			}
		}
	}
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			err = a.insertPolicyLine(&line)
			if err != nil {
				return
			}
		}
	}
	return
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) (err error) {
	line := savePolicyLine(ptype, rule)
	err = a.insertPolicyLine(&line)
	if err != nil {
		return
	}
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) (err error) {
	line := savePolicyLine(ptype, rule)
	err = a.deletePolicyLine(&line)
	if err != nil {
		return
	}
	return err
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) (err error) {
	line := CasbinRule{}
	line.PType = ptype
	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}
	err = a.rawDelete(&line)
	if err != nil {
		return
	}
	return
}

func (a *Adapter) rawDelete(line *CasbinRule) (err error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE ptype = ?", a.tableName) // nolint: gosec
	args := []interface{}{line.PType}

	if line.V0 != "" {
		query += " AND v0 = ?"
		args = append(args, line.V0)
	}
	if line.V1 != "" {
		query += " AND v1 = ?"
		args = append(args, line.V1)
	}
	if line.V2 != "" {
		query += " AND v2 = ?"
		args = append(args, line.V2)
	}
	if line.V3 != "" {
		query += " AND v3 = ?"
		args = append(args, line.V3)
	}
	if line.V4 != "" {
		query += " AND v4 = ?"
		args = append(args, line.V4)
	}
	if line.V5 != "" {
		query += " AND v5 = ?"
		args = append(args, line.V5)
	}
	_, err = a.db.Exec(query, args...)
	if err != nil {
		return
	}
	return
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *Adapter) IsFiltered() bool {
	return a.isFiltered
}
