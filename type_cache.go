package csvpp

import (
	"reflect"
	"sync"
)

// typeInfo holds cached reflection data for a struct type.
// It is computed once per reflect.Type and reused across all
// Marshal/Unmarshal calls for that type.
// All fields are immutable after creation, making it safe for concurrent access.
// The source of truth is struct tags, which are compile-time constants,
// so the cached values can never become stale.
// Callers must NOT mutate the returned slices (headers, encodeFields, tagNames).
type typeInfo struct {
	headers      []*ColumnHeader
	encodeFields []encodeFieldInfo
	tagNames     []tagNameInfo
}

// encodeFieldInfo maps a struct field index to its position in the headers slice.
// This eliminates per-record tag scanning in encodeRecord.
type encodeFieldInfo struct {
	structIndex int // index in the struct (for reflect.Value.Field)
	headerIndex int // index in the headers slice
}

// tagNameInfo holds the extracted tag name for a struct field,
// used by buildFieldMap to match against CSV headers without re-parsing tags.
type tagNameInfo struct {
	structIndex int
	tagName     string
}

// typeInfoCache caches typeInfo by reflect.Type.
// This follows the same pattern as encoding/json's fieldCache.
// Entries are never evicted; this is safe because the key space (distinct struct types
// used with Marshal/Unmarshal) is bounded by the application's type definitions.
var typeInfoCache sync.Map // map[reflect.Type]*typeInfo

// cachedTypeInfo returns the cached typeInfo for the given type,
// computing it on first access.
func cachedTypeInfo(t reflect.Type) *typeInfo {
	if ti, ok := typeInfoCache.Load(t); ok {
		return ti.(*typeInfo)
	}
	ti, _ := typeInfoCache.LoadOrStore(t, newTypeInfo(t))
	return ti.(*typeInfo)
}

// newTypeInfo computes typeInfo for a struct type.
// This consolidates the tag-scanning logic from buildHeaders,
// buildFieldMap, and encodeRecord into a single pass.
func newTypeInfo(t reflect.Type) *typeInfo {
	ti := &typeInfo{}

	headerIdx := 0
	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("csvpp")
		if tag == "" || tag == "-" {
			continue
		}

		// Parse header from tag.
		h, err := parseColumnHeader(tag)
		if err != nil {
			h = &ColumnHeader{
				Name: tag,
				Kind: SimpleField,
			}
		}
		ti.headers = append(ti.headers, h)

		// Store encode mapping.
		ti.encodeFields = append(ti.encodeFields, encodeFieldInfo{
			structIndex: i,
			headerIndex: headerIdx,
		})

		// Store tag name for decode mapping.
		ti.tagNames = append(ti.tagNames, tagNameInfo{
			structIndex: i,
			tagName:     extractTagName(tag),
		})

		headerIdx++
	}

	return ti
}
