package fixedwidth

import (
	"reflect"
	"time"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"bytes"
)

type Unmarshaler interface {
	UnmarshalFixed([]byte) error
}

func Unmarshal(data []byte, out interface{}) (err error) {
	_, err = unmarshalRecursive(data, nil, reflect.ValueOf(out))
	return
}

func unmarshalRecursive(data []byte, field *reflect.StructField, val reflect.Value) (valid bool, err error) {
	var tagz *fixedTags
	if field != nil {
		tagz, err = parseTags(*field, val.Kind())
		if err != nil {
			return
		}
	}

	valid = true
	//custom unmarshaler
	if _, ok := val.Interface().(Unmarshaler); ok {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		err = val.Interface().(Unmarshaler).UnmarshalFixed(data)
		return
	}
	switch val.Kind() {
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		unwrapped := val.Elem()

		// Check if the pointer is nil
		if !unwrapped.IsValid() {
			newInst := reflect.New(val.Type().Elem())
			if valid, err = unmarshalRecursive(data, field, newInst); err != nil {
				return
			} else if valid {
				val.Set(newInst)
			}
			return
		}

		return unmarshalRecursive(data, field, unwrapped)
	case reflect.Interface:
		unwrapped := val.Elem()
		return unmarshalRecursive(data, field, unwrapped)
	case reflect.Struct:
		// struct type exceptions
		if t, ok := val.Interface().(time.Time); ok {
			if tagz.Format != "" {
				s := strings.Trim(string(data), tagz.Pad)
				if len(s) > 0 && s[0] != 0x00 {
					t, err = time.Parse(tagz.Format, s)
					if err != nil {
						return
					}
					val.Set(reflect.ValueOf(t))
				} else {
					valid = false
				}
			} else {
				err = errors.New("no date format specified")
				return
			}
			return
		}
		// else walk the fields
		tipe := reflect.TypeOf(val.Interface())
		pos := 0
		for i := 0; i < val.NumField(); i += 1 {
			field := tipe.Field(i)
			tagz, err = parseTags(field, field.Type.Kind())
			if err != nil {
				return
			}
			if tagz == nil {
				continue
			}
			if _, err = unmarshalRecursive(data[pos:pos+tagz.Len], &field, val.Field(i)); err != nil {
				return
			}
			pos += tagz.Len
		}
	case reflect.String:
		s := string(data)
		s = strings.Trim(s, tagz.Pad)
		val.SetString(s)
		if s == "" {
			valid = false
		}
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if len(data) > 0 && data[0] != 0x00 {
			var tmpInt int64
			if tagz.Pad != defaultPadInt {
				data = bytes.Trim(data, tagz.Pad)
			}
			if string(data) == "" {
				valid = false
				return
			}
			tmpInt, err = strconv.ParseInt(string(data), tagz.Base, 64)
			if err != nil {
				err = errors.New(fmt.Sprintf("parseInt error for field %s tag %s, %s", field.Name, field.Tag.Get(tagName), err.Error()))
				return
			}
			val.SetInt(tmpInt)
		}
	case reflect.Slice:
		if _, ok := val.Interface().([]byte); ok {
			val.SetBytes(data)
			return
		}
		err = errors.New(fmt.Sprintf("Unknown slice type %s", val.Kind()))
		return

	default:
		err = errors.New("Unknown type: " + val.Kind().String())
		return
	}
	return
}
