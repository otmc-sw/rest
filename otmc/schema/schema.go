/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package schema

import (
	"reflect"
	"strings"
	"time"
)

var (
	timeType  = reflect.TypeOf(time.Time{})
	byteSlice = reflect.TypeOf([]byte(nil))
)

// Field describes a single column/API field derived from a Go struct field.
type Field struct {
	GoName   string       // Go struct field name, e.g. "RoleID"
	JSONName string       // JSON property name, e.g. "role_id"
	DBName   string       // Database column name, e.g. "role_id"
	Type     reflect.Type // Go type (dereferenced for pointers)
	Kind     reflect.Kind
	Nullable bool // pointer type or sql.Null* type

	Primary  bool
	Auto     bool // auto-increment primary key
	Required bool
	Unique   bool
	Index    bool
	Hidden   bool // never exposed in API responses
	ReadOnly bool // accepted on read but ignored on create/update input
	Size     int  // max length for string-like columns, 0 = default
}

// Relation describes an association to another registered model.
type Relation struct {
	Name      string // Go field name, e.g. "Role"
	Type      reflect.Type
	Kind      string // "belongs_to" | "has_many" | "has_one"
	ForeignKey string // column on this model holding the FK, if any
}

// Model is the parsed metadata of one registered Go struct.
type Model struct {
	GoType   reflect.Type
	Name     string // singular type name, e.g. "User"
	Table    string // table name, e.g. "users"
	Route    string // route segment, e.g. "users"
	Fields   []*Field
	ByJSON   map[string]*Field
	ByDB     map[string]*Field
	Primary  *Field
	Relations []*Relation
}

// Parse extracts OTMC metadata from a pointer to a struct (e.g. &models.User{}).
func Parse(model any) (*Model, error) {
	t := reflect.TypeOf(model)
	if t == nil || t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return nil, errInvalidModel(t)
	}
	t = t.Elem()
	m := &Model{
		GoType:  t,
		Name:    t.Name(),
		Table:   Plural(Snake(t.Name())),
		Route:   Plural(Snake(t.Name())),
		ByJSON:  map[string]*Field{},
		ByDB:    map[string]*Field{},
	}
	m.Table = resolveTableName(t, m.Table)
	parseFields(t, m, "")
	if m.Primary == nil {
		// Convention: field named "ID" of integer kind is the primary key.
		for _, f := range m.Fields {
			if f.GoName == "ID" && isIntegerKind(f.Kind) {
				f.Primary, f.Auto = true, true
				m.Primary = f
				break
			}
		}
	}
	return m, nil
}

func parseFields(t reflect.Type, m *Model, prefix string) {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		otmcTag := sf.Tag.Get("otmc")
		if otmcTag == "-" || otmcTag == "ignore" {
			continue
		}
		jsonTag := sf.Tag.Get("json")
		jsonName, _ := parseJSONTag(jsonTag)
		if jsonName == "-" && otmcTag == "" {
			continue // json:"-" hides the field entirely unless otmc metadata exists
		}
		ft := sf.Type
		kind := ft.Kind()
		if kind == reflect.Ptr {
			ft = ft.Elem()
			kind = ft.Kind()
		}

		// Anonymous embedded structs: flatten their fields.
		if sf.Anonymous && kind == reflect.Struct && ft != timeType && !isSQLNull(ft) {
			parseFields(ft, m, prefix)
			continue
		}

		// Associations: struct kinds other than time.Time / sql.Null* are relations.
		if kind == reflect.Struct && ft != timeType && !isSQLNull(ft) {
			m.Relations = append(m.Relations, &Relation{
				Name: sf.Name,
				Type: ft,
				Kind: relationKind(sf, kind, t),
			})
			continue
		}
		if kind == reflect.Slice && ft.Elem().Kind() == reflect.Struct {
			m.Relations = append(m.Relations, &Relation{
				Name: sf.Name,
				Type: ft.Elem(),
				Kind: "has_many",
			})
			continue
		}

		directives := parseDirectives(otmcTag)
		f := &Field{
			GoName:   sf.Name,
			JSONName: jsonName,
			DBName:   Snake(sf.Name),
			Type:     ft,
			Kind:     kind,
			Nullable: sf.Type.Kind() == reflect.Ptr || isSQLNull(ft),
		}
		if f.JSONName == "" {
			f.JSONName = f.DBName
		}
		f.Primary = has(directives, "primary")
		f.Auto = has(directives, "auto") || f.Primary && isIntegerKind(kind)
		f.Required = has(directives, "required")
		f.Unique = has(directives, "unique")
		f.Index = has(directives, "index")
		f.Hidden = has(directives, "hidden")
		f.ReadOnly = has(directives, "readonly") || has(directives, "read_only")
		if s, ok := directives["size"]; ok {
			f.Size = atoi(s)
		}
		if f.Primary {
			m.Primary = f
		}
		m.Fields = append(m.Fields, f)
		m.ByJSON[f.JSONName] = f
		m.ByDB[f.DBName] = f
	}
}

func relationKind(sf reflect.StructField, kind reflect.Kind, owner reflect.Type) string {
	// *Role or Role -> belongs_to; heuristics only, FK resolution happens in store.
	_ = owner
	return "belongs_to"
}

// resolveTableName honors a `TableName() string` method on the model (pointer receiver).
func resolveTableName(t reflect.Type, def string) string {
	if reflect.PtrTo(t).Implements(reflect.TypeOf((*interface{ TableName() string })(nil)).Elem()) {
		if name := reflect.New(t).Interface().(interface{ TableName() string }).TableName(); name != "" {
			return name
		}
	}
	return def
}

func has(directives map[string]string, key string) bool {
	_, ok := directives[key]
	return ok
}

func isSQLNull(t reflect.Type) bool {
	return t.PkgPath() == "database/sql" && strings.HasPrefix(t.Name(), "Null")
}

func isIntegerKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
