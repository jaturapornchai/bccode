// Package textguard finds text PostgreSQL cannot store.
package textguard

import (
	"fmt"
	"reflect"
	"strings"
)

// NULField returns the JSON path (for example "position" or "names[1].name") of the first
// string in v that contains U+0000, or "" when none does. PostgreSQL TEXT and JSONB reject
// that byte (SQLSTATE 22021), so a save carrying it — usually text pasted from a PDF — failed
// as a generic "try again" 500 instead of naming the field to fix.
func NULField(v any) string {
	return nulPath(reflect.ValueOf(v), "")
}

// Message is the user-facing text for a NUL found at path: the field label (labels maps the last
// path segment to a languages.tsv key) inside the text_contains_nul_field row, or the
// text_contains_nul row when the field has no label. text renders a languages.tsv key.
func Message(path string, labels map[string]string, text func(key string) string) string {
	if key, ok := labels[LastSegment(path)]; ok {
		return strings.ReplaceAll(text("text_contains_nul_field"), "{0}", text(key))
	}
	return text("text_contains_nul")
}

// LastSegment is the field name of a path without its parents or index: "names[1].name" → "name".
func LastSegment(path string) string {
	if i := strings.LastIndexByte(path, '.'); i >= 0 {
		path = path[i+1:]
	}
	if i := strings.IndexByte(path, '['); i >= 0 {
		path = path[:i]
	}
	return path
}

func nulPath(v reflect.Value, path string) string {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return ""
		}
		return nulPath(v.Elem(), path)
	case reflect.String:
		if strings.IndexByte(v.String(), 0) >= 0 {
			return path
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			if !f.IsExported() || tag == "-" {
				continue
			}
			name, _, _ := strings.Cut(tag, ",")
			child := path
			if !(f.Anonymous && name == "") {
				if name == "" {
					name = f.Name
				}
				child = strings.TrimPrefix(path+"."+name, ".")
			}
			if found := nulPath(v.Field(i), child); found != "" {
				return found
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if found := nulPath(v.Index(i), fmt.Sprintf("%s[%d]", path, i)); found != "" {
				return found
			}
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprint(iter.Key().Interface())
			if strings.IndexByte(key, 0) >= 0 {
				return path
			}
			if found := nulPath(iter.Value(), strings.TrimPrefix(path+"."+key, ".")); found != "" {
				return found
			}
		}
	}
	return ""
}
