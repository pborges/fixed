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
				switch v.Field(i).Kind() {
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
						v.Field(i).SetInt(tmpInt)
					}
				case reflect.String:
					var pad = defaultPadString
					if p, ok := tags[tagPad]; ok {
						pad = p
					}
					if len(buff) > 0 && buff[0] != 0x00 {
						v.Field(i).SetString(strings.Trim(string(buff), pad))
					}
				case reflect.Slice:
					if _, ok := v.Field(i).Interface().([]byte); ok {
						v.Field(i).SetBytes(buff)
					} else {
						err = errors.New(fmt.Sprintf("Unknown slice type %s", v.Field(i).Kind()))
						return
					}
				case reflect.Struct:
					if t, ok := v.Field(i).Interface().(time.Time); ok {
						if tag, ok := tags[tagFormat]; ok {
							t, err = time.Parse(tag, string(buff))
							if err != nil {
								return
							}
							v.Field(i).Set(reflect.ValueOf(t))
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
