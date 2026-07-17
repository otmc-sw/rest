/**
 * @License Apache License 2.0
 * @Copyright (c) 2026 OTMC Softwares. OTMC Golang REST.
 * @Contributors Nguyen Van Trung, Nguyen Thi Hoai, OTMC Contributors.
**/
package mapper

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMapPointerStringToString(t *testing.T) {
	type Src struct {
		Name *string
	}

	type Dst struct {
		Name string
	}

	str := "hello"
	src := Src{Name: &str}
	dst := Map[Dst, Src](src)

	if dst.Name != "hello" {
		t.Fatalf("expected Name='hello', got %v", dst.Name)
	}
}

func TestMapPointerStringToStringNil(t *testing.T) {
	type Src struct {
		Name *string
	}

	type Dst struct {
		Name string
	}

	src := Src{Name: nil}
	dst := Map[Dst, Src](src)

	if dst.Name != "" {
		t.Fatalf("expected Name='', got %v", dst.Name)
	}
}

func TestMapPointerBoolToBool(t *testing.T) {
	type Src struct {
		Active *bool
	}

	type Dst struct {
		Active bool
	}

	active := true
	src := Src{Active: &active}
	dst := Map[Dst, Src](src)

	if dst.Active != true {
		t.Fatalf("expected Active=true, got %v", dst.Active)
	}
}

func TestMapPointerBoolToBoolNil(t *testing.T) {
	type Src struct {
		Active *bool
	}

	type Dst struct {
		Active bool
	}

	src := Src{Active: nil}
	dst := Map[Dst, Src](src)

	if dst.Active != false {
		t.Fatalf("expected Active=false, got %v", dst.Active)
	}
}

func TestMapPointerInt64ToInt64(t *testing.T) {
	type Src struct {
		Age *int64
	}

	type Dst struct {
		Age int64
	}

	age := int64(25)
	src := Src{Age: &age}
	dst := Map[Dst, Src](src)

	if dst.Age != 25 {
		t.Fatalf("expected Age=25, got %v", dst.Age)
	}
}

func TestMapPointerInt64ToInt64Nil(t *testing.T) {
	type Src struct {
		Age *int64
	}

	type Dst struct {
		Age int64
	}

	src := Src{Age: nil}
	dst := Map[Dst, Src](src)

	if dst.Age != 0 {
		t.Fatalf("expected Age=0, got %v", dst.Age)
	}
}

func TestMapPointerFloat64ToFloat64(t *testing.T) {
	type Src struct {
		Score *float64
	}

	type Dst struct {
		Score float64
	}

	score := float64(98.5)
	src := Src{Score: &score}
	dst := Map[Dst, Src](src)

	if dst.Score != 98.5 {
		t.Fatalf("expected Score=98.5, got %v", dst.Score)
	}
}

func TestMapPointerFloat64ToFloat64Nil(t *testing.T) {
	type Src struct {
		Score *float64
	}

	type Dst struct {
		Score float64
	}

	src := Src{Score: nil}
	dst := Map[Dst, Src](src)

	if dst.Score != 0 {
		t.Fatalf("expected Score=0, got %v", dst.Score)
	}
}

func TestMapPointerTimeToTime(t *testing.T) {
	type Src struct {
		Created *time.Time
	}

	type Dst struct {
		Created time.Time
	}

	now := time.Now()
	src := Src{Created: &now}
	dst := Map[Dst, Src](src)

	if !dst.Created.Equal(now) {
		t.Fatalf("expected Created=%v, got %v", now, dst.Created)
	}
}

func TestMapPointerTimeToTimeNil(t *testing.T) {
	type Src struct {
		Created *time.Time
	}

	type Dst struct {
		Created time.Time
	}

	src := Src{Created: nil}
	dst := Map[Dst, Src](src)

	if !dst.Created.IsZero() && !dst.Created.Equal(time.Time{}) {
		t.Fatalf("expected zero time, got %v", dst.Created)
	}
}

func TestMapSliceToSlice(t *testing.T) {
	type Src struct {
		Tags *[]string
	}

	type Dst struct {
		Tags []string
	}

	tags := []string{"go", "rest"}
	src := Src{Tags: &tags}
	dst := Map[Dst, Src](src)

	if len(dst.Tags) != 2 || dst.Tags[0] != "go" || dst.Tags[1] != "rest" {
		t.Fatalf("expected Tags=['go','rest'], got %v", dst.Tags)
	}
}

func TestMapSliceToSliceNil(t *testing.T) {
	type Src struct {
		Tags *[]string
	}

	type Dst struct {
		Tags []string
	}

	src := Src{Tags: nil}
	dst := Map[Dst, Src](src)

	if dst.Tags != nil {
		t.Fatalf("expected Tags=nil, got %v", dst.Tags)
	}
}

func TestMapMapToMap(t *testing.T) {
	type Src struct {
		Meta *map[string]int64
	}

	type Dst struct {
		Meta map[string]int64
	}

	meta := map[string]int64{"a": 1}
	src := Src{Meta: &meta}
	dst := Map[Dst, Src](src)

	if dst.Meta["a"] != 1 {
		t.Fatalf("expected Meta['a']=1, got %v", dst.Meta)
	}
}

func TestMapMapToMapNil(t *testing.T) {
	type Src struct {
		Meta *map[string]int64
	}

	type Dst struct {
		Meta map[string]int64
	}

	src := Src{Meta: nil}
	dst := Map[Dst, Src](src)

	if dst.Meta != nil {
		t.Fatalf("expected Meta=nil, got %v", dst.Meta)
	}
}

func TestMapStructToStruct(t *testing.T) {
	type InnerSrc struct {
		Name string
	}

	type Src struct {
		User *InnerSrc
	}

	type InnerDst struct {
		Name string
	}

	type Dst struct {
		User InnerDst
	}

	src := Src{User: &InnerSrc{Name: "John"}}
	dst := Map[Dst, Src](src)

	if dst.User.Name != "John" {
		t.Fatalf("expected User.Name='John', got %v", dst.User.Name)
	}
}

func TestMapStructToStructNil(t *testing.T) {
	type InnerSrc struct {
		Name string
	}

	type Src struct {
		User *InnerSrc
	}

	type InnerDst struct {
		Name string
	}

	type Dst struct {
		User InnerDst
	}

	src := Src{User: nil}
	dst := Map[Dst, Src](src)

	if dst.User.Name != "" {
		t.Fatalf("expected User.Name='', got %v", dst.User.Name)
	}
}

func TestMapCollectionToNullString(t *testing.T) {
	type Src struct {
		Tags []string
	}

	type Dst struct {
		Tags string
	}

	src := Src{Tags: []string{"go", "rest"}}
	dst := Map[Dst, Src](src)

	if dst.Tags == "" {
		t.Fatalf("expected Tags JSON string, got empty")
	}
}

func TestMapJSONToNullString(t *testing.T) {
	type Src struct {
		Data json.RawMessage
	}

	type Dst struct {
		Data string
	}

	src := Src{Data: json.RawMessage(`{"key":"value"}`)}
	dst := Map[Dst, Src](src)

	if dst.Data == "" {
		t.Fatalf("expected Data JSON string, got empty")
	}
}

func TestMapNullStringToJSON(t *testing.T) {
	type Src struct {
		Data string
	}

	type Dst struct {
		Data json.RawMessage
	}

	src := Src{Data: `{"key":"value"}`}
	dst := Map[Dst, Src](src)

	if string(dst.Data) != `{"key":"value"}` {
		t.Fatalf("expected json.RawMessage, got %v", string(dst.Data))
	}
}