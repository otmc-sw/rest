/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/otmc-sw/rest/app/schema"
)

// Query carries list/search options parsed from HTTP query params.
type Query struct {
	Page    int
	Size    int
	Keyword string
	Sort    string
	Order   string            // "asc" | "desc"
	Filters map[string]string // field json/db name -> raw value
}

// PageResult is a paginated list response.
type PageResult struct {
	Items      []any
	Total      int64
	Page       int
	Size       int
	TotalPages int64
}

// Repository provides generic CRUD over one registered model.
type Repository struct {
	db    *gorm.DB
	model *schema.Model
}

func NewRepository(db *gorm.DB, model *schema.Model) *Repository {
	return &Repository{db: db, model: model}
}

// Model returns the schema metadata.
func (r *Repository) Model() *schema.Model { return r.model }

func (r *Repository) newInstance() any {
	return reflect.New(r.model.GoType).Interface()
}

func (r *Repository) base(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Table(r.model.Table)
}

// List returns a paginated, filtered, sorted page of records.
func (r *Repository) List(ctx context.Context, q Query) (*PageResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 {
		q.Size = 20
	}
	if q.Size > 100 {
		q.Size = 100
	}

	countQ, err := r.applyQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	var total int64
	if res := countQ.Count(&total); res.Error != nil {
		return nil, res.Error
	}

	findQ, err := r.applyQuery(ctx, q)
	if err != nil {
		return nil, err
	}
	findQ = findQ.Offset((q.Page - 1) * q.Size).Limit(q.Size)
	if order, ok := r.sortClause(q); ok {
		findQ = findQ.Order(order)
	} else if r.model.Primary != nil {
		findQ = findQ.Order(r.model.Primary.DBName + " ASC")
	}

	slice := reflect.New(reflect.SliceOf(r.model.GoType))
	if res := findQ.Find(slice.Interface()); res.Error != nil {
		return nil, res.Error
	}

	items := make([]any, 0, slice.Elem().Len())
	for i := 0; i < slice.Elem().Len(); i++ {
		items = append(items, slice.Elem().Index(i).Addr().Interface())
	}
	totalPages := int64(0)
	if q.Size > 0 {
		totalPages = (total + int64(q.Size) - 1) / int64(q.Size)
	}
	return &PageResult{Items: items, Total: total, Page: q.Page, Size: q.Size, TotalPages: totalPages}, nil
}

// Get returns one record by primary key.
func (r *Repository) Get(ctx context.Context, id any) (any, error) {
	inst := r.newInstance()
	if res := r.base(ctx).Where(r.pkCond(), id).First(inst); res.Error != nil {
		return nil, res.Error
	}
	return inst, nil
}

// Create validates and inserts a record decoded from JSON.
func (r *Repository) Create(ctx context.Context, body []byte) (any, error) {
	inst := r.newInstance()
	if err := json.Unmarshal(body, inst); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	if err := r.applyTransforms(inst); err != nil {
		return nil, err
	}
	if err := r.validate(inst); err != nil {
		return nil, err
	}
	if err := r.checkUnique(ctx, inst, nil); err != nil {
		return nil, err
	}
	if res := r.base(ctx).Create(inst); res.Error != nil {
		return nil, res.Error
	}
	return inst, nil
}

// Update fully replaces a record (PUT semantics).
func (r *Repository) Update(ctx context.Context, id any, body []byte) (any, error) {
	existing, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, existing); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	r.setPrimary(existing, id)
	if err := r.applyTransforms(existing); err != nil {
		return nil, err
	}
	if err := r.validate(existing); err != nil {
		return nil, err
	}
	if err := r.checkUnique(ctx, existing, id); err != nil {
		return nil, err
	}
	if res := r.base(ctx).Save(existing); res.Error != nil {
		return nil, res.Error
	}
	return existing, nil
}

// Patch applies only the fields present in the JSON body (PATCH semantics).
func (r *Repository) Patch(ctx context.Context, id any, body []byte) (any, error) {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	return r.patchRecord(ctx, id, raw)
}

// patchRecord applies a decoded partial update to a single record.
func (r *Repository) patchRecord(ctx context.Context, id any, raw map[string]json.RawMessage) (any, error) {
	existing, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	ev := reflect.ValueOf(existing).Elem()
	for key, val := range raw {
		f := r.model.ByJSON[key]
		if f == nil {
			f = r.model.ByDB[key]
		}
		if f == nil || f.Primary || f.ReadOnly {
			continue
		}
		fv := ev.FieldByName(f.GoName)
		if !fv.CanSet() {
			continue
		}
		if fv.Kind() == reflect.Ptr {
			if string(val) == "null" {
				fv.Set(reflect.Zero(fv.Type()))
				continue
			}
			pv := reflect.New(fv.Type().Elem())
			if err := json.Unmarshal(val, pv.Interface()); err != nil {
				return nil, fmt.Errorf("invalid value for %q: %w", key, err)
			}
			fv.Set(pv)
			continue
		}
		if err := json.Unmarshal(val, fv.Addr().Interface()); err != nil {
			return nil, fmt.Errorf("invalid value for %q: %w", key, err)
		}
	}
	r.setPrimary(existing, id)
	if err := r.applyTransforms(existing); err != nil {
		return nil, err
	}
	if err := r.validate(existing); err != nil {
		return nil, err
	}
	if err := r.checkUnique(ctx, existing, id); err != nil {
		return nil, err
	}
	if res := r.base(ctx).Save(existing); res.Error != nil {
		return nil, res.Error
	}
	return existing, nil
}

// UpdateMany applies PATCH semantics to every record in ids.
// Returns the number of records updated; stops at the first error.
func (r *Repository) UpdateMany(ctx context.Context, ids []any, body []byte) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, fmt.Errorf("invalid request body: %w", err)
	}
	var updated int64
	for _, id := range ids {
		if _, err := r.patchRecord(ctx, id, raw); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

// UpdateManyItems applies PATCH semantics per record. Each item must be a JSON
// object containing the record's primary key (by JSON name) plus the fields to
// update, e.g. [{"id": 1, "name": "a"}, {"id": 2, "name": "b"}].
// Returns the number of records updated; stops at the first error.
func (r *Repository) UpdateManyItems(ctx context.Context, items []json.RawMessage) (int64, error) {
	if len(items) == 0 {
		return 0, nil
	}
	key := "id"
	if r.model.Primary != nil && r.model.Primary.JSONName != "" {
		key = r.model.Primary.JSONName
	}
	var updated int64
	for _, item := range items {
		raw := map[string]json.RawMessage{}
		if err := json.Unmarshal(item, &raw); err != nil {
			return updated, fmt.Errorf("invalid update item: %w", err)
		}
		idRaw, ok := raw[key]
		if !ok {
			return updated, fmt.Errorf("update item missing primary key %q", key)
		}
		id, err := r.ParseID(rawToString(idRaw))
		if err != nil {
			return updated, err
		}
		if _, err := r.patchRecord(ctx, id, raw); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

// rawToString converts a JSON scalar (string or number) to its plain string
// form so it can be fed to ParseID.
func rawToString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// Delete removes one record by primary key.
func (r *Repository) Delete(ctx context.Context, id any) error {
	res := r.base(ctx).Where(r.pkCond(), id).Delete(r.newInstance())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteMany removes records by a list of primary keys.
func (r *Repository) DeleteMany(ctx context.Context, ids []any) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := r.base(ctx).Where(r.pkCond()+" IN ?", ids).Delete(r.newInstance())
	return res.RowsAffected, res.Error
}

// ParseID converts a string path param into the primary key Go type.
func (r *Repository) ParseID(s string) (any, error) {
	if r.model.Primary == nil {
		return nil, fmt.Errorf("otmc: model %s has no primary key", r.model.Name)
	}
	return parseScalar(s, r.model.Primary.Type, r.model.Primary.Kind)
}

func (r *Repository) pkCond() string {
	if r.model.Primary != nil {
		return r.quote(r.model.Primary.DBName)
	}
	return r.quote("id")
}

func (r *Repository) setPrimary(inst any, id any) {
	if r.model.Primary == nil {
		return
	}
	ev := reflect.ValueOf(inst).Elem()
	fv := ev.FieldByName(r.model.Primary.GoName)
	if fv.IsValid() && fv.CanSet() {
		fv.Set(reflect.ValueOf(id))
	}
}

// applyQuery builds a filtered gorm.DB for List count/find.
func (r *Repository) applyQuery(ctx context.Context, q Query) (*gorm.DB, error) {
	db := r.base(ctx).Model(r.newInstance())
	for key, raw := range q.Filters {
		f := r.fieldByName(key)
		if f == nil {
			continue
		}
		val, err := parseScalar(raw, f.Type, f.Kind)
		if err != nil {
			return nil, fmt.Errorf("invalid filter %q: %w", key, err)
		}
		db = db.Where(r.quote(f.DBName)+" = ?", val)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		clauses := []string{}
		args := []any{}
		for _, f := range r.model.Fields {
			if f.Kind == reflect.String {
				clauses = append(clauses, r.quote(f.DBName)+" LIKE ?")
				args = append(args, like)
			}
		}
		if len(clauses) > 0 {
			db = db.Where("("+strings.Join(clauses, " OR ")+")", args...)
		}
	}
	return db, nil
}

func (r *Repository) fieldByName(key string) *schema.Field {
	if f := r.model.ByJSON[key]; f != nil {
		return f
	}
	return r.model.ByDB[key]
}

func (r *Repository) sortClause(q Query) (string, bool) {
	f := r.fieldByName(q.Sort)
	if f == nil {
		return "", false
	}
	order := "ASC"
	if strings.EqualFold(q.Order, "desc") {
		order = "DESC"
	}
	return r.quote(f.DBName) + " " + order, true
}

// quote returns the dialect-quoted column name so reserved words (e.g.
// "group", "key") work as filters, sort keys and primary-key conditions.
func (r *Repository) quote(name string) string {
	stmt := &gorm.Statement{DB: r.db}
	return stmt.Quote(name)
}

// validate enforces required fields on write.
func (r *Repository) validate(inst any) error {
	ev := reflect.ValueOf(inst).Elem()
	for _, f := range r.model.Fields {
		if !f.Required {
			continue
		}
		fv := ev.FieldByName(f.GoName)
		if !fv.IsValid() {
			continue
		}
		if isZero(fv) {
			return &ValidationError{Field: f.JSONName}
		}
	}
	return nil
}

// checkUnique enforces app-level uniqueness for fields tagged `unique`.
func (r *Repository) checkUnique(ctx context.Context, inst any, excludeID any) error {
	ev := reflect.ValueOf(inst).Elem()
	for _, f := range r.model.Fields {
		if !f.Unique {
			continue
		}
		fv := ev.FieldByName(f.GoName)
		if !fv.IsValid() || isZero(fv) {
			continue
		}
		q := r.base(ctx).Model(r.newInstance()).Where(r.quote(f.DBName)+" = ?", fv.Interface())
		if excludeID != nil && r.model.Primary != nil {
			q = q.Where("NOT ("+r.quote(r.model.Primary.DBName)+" = ?)", excludeID)
		}
		var count int64
		if res := q.Count(&count); res.Error != nil {
			return res.Error
		}
		if count > 0 {
			return &UniqueError{Field: f.JSONName}
		}
	}
	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return strings.TrimSpace(v.String()) == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		return false
	}
	return false
}

// parseScalar converts a raw string into a correctly typed Go value.
func parseScalar(raw string, t reflect.Type, kind reflect.Kind) (any, error) {
	v := reflect.New(t).Elem()
	if kind == reflect.String {
		v.SetString(raw)
		return v.Interface(), nil
	}
	if t.PkgPath() == "time" && t.Name() == "Time" {
		for _, layout := range []string{time.RFC3339, "2006-01-02", "2006-01-02 15:04:05"} {
			if tm, err := time.Parse(layout, raw); err == nil {
				v.Set(reflect.ValueOf(tm))
				return v.Interface(), nil
			}
		}
		return nil, fmt.Errorf("invalid time %q", raw)
	}
	var err error
	switch kind {
	case reflect.Bool:
		var b bool
		b, err = strconv.ParseBool(raw)
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		n, err = strconv.ParseInt(raw, 10, 64)
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var n uint64
		n, err = strconv.ParseUint(raw, 10, 64)
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		var f float64
		f, err = strconv.ParseFloat(raw, 64)
		v.SetFloat(f)
	default:
		return raw, nil
	}
	if err != nil {
		return nil, err
	}
	return v.Interface(), nil
}
