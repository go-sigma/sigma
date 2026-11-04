// Copyright 2026 sigma
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

package config

import (
	"encoding"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// envPrefix is the prefix every configuration environment variable carries.
	envPrefix = "SIGMA_"
	// envSeparator splits list values, for example daemon.builder.extra_hosts.
	envSeparator = ","
)

// envField is one configuration leaf that can be overridden from the environment.
type envField struct {
	// name is the canonical variable name, for example SIGMA_DAEMON_BUILDER_TLSVERIFY.
	name string
	// key is the normalized lookup key shared by every accepted spelling.
	key string
	// path is the YAML path used in logs and error messages, for example daemon.builder.tlsVerify.
	path string
	// index locates the field inside Configuration.
	index []int
}

// envCatalog splits every configuration leaf into the ones that a single
// environment variable can express and the ones that cannot, because they are
// structured values such as lists of objects.
type envCatalog struct {
	supported []envField
	skipped   []envField
}

var (
	catalogOnce = sync.OnceValue(loadEnvCatalog)
	// textUnmarshalerType lets custom leaf types keep control over their own parsing.
	textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()
	// validatorType is implemented by the generated enums, so unknown enum
	// values are rejected instead of silently stored.
	validatorType = reflect.TypeFor[interface{ IsValid() bool }]()
	durationType  = reflect.TypeFor[time.Duration]()
	timeType      = reflect.TypeFor[time.Time]()
)

func loadEnvCatalog() envCatalog {
	catalog := envCatalog{supported: []envField{}, skipped: []envField{}}
	catalog.collect(reflect.TypeFor[Configuration](), nil, nil)
	catalog.supported = dedupeEnvFields(catalog.supported)
	catalog.skipped = dedupeEnvFields(catalog.skipped)
	return catalog
}

// collect walks one struct type and records every leaf as an env field.
func (c *envCatalog) collect(typ reflect.Type, path []string, index []int) {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := yamlFieldName(field)
		if name == "" {
			continue
		}
		fieldPath := append(append(make([]string, 0, len(path)+1), path...), name)
		fieldIndex := append(append(make([]int, 0, len(index)+1), index...), i)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		entry := newEnvField(fieldPath, fieldIndex)
		switch {
		case fieldType.Kind() == reflect.Struct && !isScalarStruct(fieldType):
			c.collect(fieldType, fieldPath, fieldIndex)
		case fieldType.Kind() == reflect.Slice && !isScalarType(fieldType.Elem()):
			c.skipped = append(c.skipped, entry)
		case fieldType.Kind() == reflect.Map:
			c.skipped = append(c.skipped, entry)
		default:
			c.supported = append(c.supported, entry)
		}
	}
}

// yamlFieldName returns the YAML key of a field and reports whether it should be
// skipped by returning an empty string.
func yamlFieldName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
	switch name {
	case "":
		if field.Name == "" {
			return ""
		}
		return field.Name
	case "-":
		return ""
	default:
		return name
	}
}

// isScalarStruct reports whether a struct type is a single value parsed from a string.
func isScalarStruct(typ reflect.Type) bool {
	return typ == durationType || typ == timeType
}

// isScalarType reports whether a type can be parsed from one comma separated token.
func isScalarType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func newEnvField(path []string, index []int) envField {
	name := envName(path)
	return envField{
		name:  name,
		key:   envLookupKey(name),
		path:  strings.Join(path, "."),
		index: index,
	}
}

// envName builds the canonical variable name of a YAML path.
func envName(path []string) string {
	return envPrefix + strings.ToUpper(strings.Join(path, "_"))
}

// envLookupKey normalizes a variable name so that SIGMA_DAEMON_BUILDER_TLSVERIFY
// and SIGMA_DAEMON_BUILDER_TLS_VERIFY are the same lookup key.
func envLookupKey(name string) string {
	return strings.ToUpper(strings.NewReplacer("_", "", "-", "").Replace(name))
}

// dedupeEnvFields keeps the first field of every lookup key so that two YAML
// paths that normalize to the same variable name do not fight over it.
func dedupeEnvFields(fields []envField) []envField {
	seen := make(map[string]struct{}, len(fields))
	result := make([]envField, 0, len(fields))
	for _, field := range fields {
		if _, ok := seen[field.key]; ok {
			slog.Error("configuration env variable collides with another item, keeping the first one",
				"env", field.name, "path", field.path)
			continue
		}
		seen[field.key] = struct{}{}
		result = append(result, field)
	}
	return result
}

// ApplyEnvOverrides applies environment variable overrides after YAML loading.
//
// Every leaf of Configuration can be set with SIGMA_ followed by the YAML path,
// joined by underscores and upper cased, for example SIGMA_DAEMON_BUILDER_ENABLED
// for daemon.builder.enabled. Underscores inside a path segment are optional, so
// SIGMA_DAEMON_BUILDER_TLS_VERIFY also works. An override only replaces values
// already loaded from the config file; unset variables keep the file value.
func (c *Configuration) ApplyEnvOverrides() error {
	return c.applyEnvOverrides(os.Environ())
}

// applyEnvOverrides is ApplyEnvOverrides with an injectable environment, so tests
// do not have to mutate the process environment.
func (c *Configuration) applyEnvOverrides(environ []string) error {
	catalog := catalogOnce()
	values := envValues(environ)
	root := reflect.ValueOf(c).Elem()
	var errs []error
	for _, field := range catalog.supported {
		raw, ok := values[field.key]
		if !ok {
			continue
		}
		err := setEnvValue(root.FieldByIndex(field.index), raw)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s (%s): %w", field.name, field.path, err))
		}
	}
	for _, field := range catalog.skipped {
		if _, ok := values[field.key]; ok {
			slog.Warn("configuration env variable is ignored because the item is a structured value, set it in the config file instead",
				"env", field.name, "path", field.path)
		}
	}
	return errors.Join(errs...)
}

// envValues indexes the SIGMA_ prefixed environment by normalized name.
func envValues(environ []string) map[string]string {
	values := make(map[string]string, len(environ))
	for _, entry := range environ {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		name = strings.ToUpper(name)
		if !strings.HasPrefix(name, envPrefix) {
			continue
		}
		values[envLookupKey(name)] = value
	}
	return values
}

// setEnvValue parses one environment value into a configuration leaf.
func setEnvValue(value reflect.Value, raw string) error {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}
	switch {
	case value.Type() == durationType:
		duration, err := time.ParseDuration(strings.TrimSpace(raw))
		if err != nil {
			return err
		}
		value.SetInt(int64(duration))
		return nil
	case value.CanAddr() && value.Addr().Type().Implements(textUnmarshalerType):
		return value.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(raw))
	}
	switch value.Kind() {
	case reflect.String:
		value.SetString(raw)
		if !isValidEnum(value) {
			return fmt.Errorf("invalid value %q", raw)
		}
	case reflect.Bool:
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return err
		}
		value.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(strings.TrimSpace(raw), 10, value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), value.Type().Bits())
		if err != nil {
			return err
		}
		value.SetFloat(parsed)
	case reflect.Slice:
		return setEnvSlice(value, raw)
	default:
		return fmt.Errorf("unsupported type %s", value.Type())
	}
	return nil
}

// isValidEnum reports whether a string leaf accepts the value it was just assigned.
// Plain strings and other string types without an IsValid method are always valid.
func isValidEnum(value reflect.Value) bool {
	if !value.CanAddr() || !value.Addr().Type().Implements(validatorType) {
		return true
	}
	return value.Addr().Interface().(interface{ IsValid() bool }).IsValid()
}

// setEnvSlice parses a comma separated list into a slice leaf.
func setEnvSlice(value reflect.Value, raw string) error {
	parts := strings.Split(raw, envSeparator)
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		items = append(items, item)
	}
	slice := reflect.MakeSlice(value.Type(), 0, len(items))
	for _, item := range items {
		elem := reflect.New(value.Type().Elem()).Elem()
		err := setEnvValue(elem, item)
		if err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}
	value.Set(slice)
	return nil
}
