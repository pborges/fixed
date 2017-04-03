package fixedwidth

import (
	"fmt"
	"encoding/binary"
	"bytes"
	"reflect"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Unmarshaler interface {
	UnmarshalFixed([]byte) error
}

func Unmarshal(data []byte, out interface{}) (err error) {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if reflect.TypeOf(out).Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	r := bytes.NewBuffer(data)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)
		tag := field.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		var fieldLen int
		tags := splitTags(field)
		fieldLen, err = strconv.Atoi(tags[tagLen])
		if err != nil {
			err = errors.New(fmt.Sprintf("fieldLen %s, %s", field.Name, err.Error()))
			return
		}
		buff := make([]byte, fieldLen)
		err = binary.Read(r, binary.BigEndian, buff)
		if err != nil {
			err = errors.New(fmt.Sprintf("read error field %s, %s", field.Name, err.Error()))
			return
		}
		if len(tag) > 0 {
			if _, ok := v.Field(i).Interface().(Unmarshaler); ok {
				if v.Field(i).IsNil() {
					v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
				}
				err = v.Field(i).Interface().(Unmarshaler).UnmarshalFixed(buff)
				if err != nil {
					return
				}
			} else {
				typ := t.Field(i).Type
				isPtr := false
				if typ.Kind() == reflect.Ptr {
					typ = typ.Elem()
					isPtr = true
				}
				switch typ.Kind() {
				case reflect.Int, reflect.Int32, reflect.Int64:
					var pad = defaultPadInt
					if p, ok := tags[tagPad]; ok {
						pad = p
					}
					if len(buff) > 0 && buff[0] != 0x00 {
						var base = 10
						if _, ok := tags[tagBase]; ok {
							base, err = strconv.Atoi(tags[tagBase])
							if err != nil {
								err = errors.New("Invalid integer base " + tags[tagBase])
								return
							}
						}

						var tmpInt int64
						tmpInt, err = strconv.ParseInt(strings.Trim(string(buff), pad), base, 64)
						if err != nil {
							err = errors.New(fmt.Sprintf("parseInt error for field %s tag %s, %s", field.Name, tag, err.Error()))
							return
						}
						val.SetInt(tmpInt)
					}
				case reflect.String:
					var pad = defaultPadString
					if p, ok := tags[tagPad]; ok {
						pad = p
					}
					s := strings.Trim(string(buff), pad)
					if len(s) > 0 && s[0] != 0x00 {
						if isPtr {
							e := reflect.New(v.Field(i).Type().Elem())
							e.Elem().SetString(s)
							v.Field(i).Set(e)
						} else {
							val.SetString(s)
						}
					}
				case reflect.Slice:
					if _, ok := v.Field(i).Interface().([]byte); ok {
						val.SetBytes(buff)
					} else {
						err = errors.New(fmt.Sprintf("Unknown slice type %s", v.Field(i).Kind()))
						return
					}
				case reflect.Struct:
					e := v.Field(i)
					isPtr := reflect.TypeOf(e.Interface()).Kind() == reflect.Ptr
					if isPtr {
						e = reflect.New(v.Field(i).Type().Elem())
						val = e.Elem()
					}
					if t, ok := val.Interface().(time.Time); ok {
						if format, ok := tags[tagFormat]; ok {
							s := strings.Trim(string(buff), " ")
							if len(s) > 0 && s[0] != 0x00 {
								t, err = time.Parse(format, s)
								if err != nil {
									return
								}
								val.Set(reflect.ValueOf(t))
								if isPtr {
									v.Field(i).Set(e)
								}
							}
						} else {
							err = errors.New("no date format specified")
							return
						}
					} else {
						err = errors.New(fmt.Sprintf("Unknown struct %s", v.Field(i).Kind()))
						return
					}
				default:
					err = errors.New(fmt.Sprintf("Unknown kind %s", v.Field(i).Kind()))
					return
				}
			}
		}
	}
	return
}
