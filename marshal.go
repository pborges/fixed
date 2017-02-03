package fixedwidth

import (
	"reflect"
	"strings"
	"fmt"
	"bytes"
	"strconv"
	"errors"
	"time"
)

type Marshaler interface {
	MarshalFixed() ([]byte, error)
}

func Marshal(in interface{}) (res []byte, err error) {
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)
	if reflect.TypeOf(in).Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	buf := bytes.Buffer{}

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
		if len(tag) > 0 {
			if _, ok := v.Field(i).Interface().(Marshaler); ok {
				if v.Field(i).IsNil() {
					v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
				}
				var b []byte
				b, err = v.Field(i).Interface().(Marshaler).MarshalFixed()
				if err != nil {
					return
				}
				_, err = buf.Write(b)
				if err != nil {
					return
				}
			} else {
				switch v.Field(i).Kind() {
				case reflect.Int, reflect.Int32, reflect.Int64:
					var pad = "0"
					if p, ok := tags[tagPad]; ok {
						pad = p
					}
					var base = 10
					if _, ok := tags[tagBase]; ok {
						base, err = strconv.Atoi(tags[tagBase])
						if err != nil {
							err = errors.New("Invalid integer base " + tags[tagBase])
							return
						}
					}

					strInt := leftPad2Len(strconv.FormatInt(v.Field(i).Int(), base), pad, fieldLen)

					// always do upper case for hex and stuff
					if base != 10 {
						strInt = strings.ToUpper(strInt)
					}

					_, err = buf.WriteString(strInt)
					if err != nil {
						return
					}
				case reflect.String:
					var pad = " "
					if p, ok := tags[tagPad]; ok {
						pad = p
					}
					strInt := leftPad2Len(v.Field(i).String(), pad, fieldLen)
					_, err = buf.WriteString(strInt)
					if err != nil {
						return
					}
				case reflect.Slice:
					if _, ok := v.Field(i).Interface().([]byte); ok {
						pad := []byte{0x0}
						if p, ok := tags[tagPad]; ok {
							pad = []byte(p)
						}
						b := bytes.Repeat(pad, fieldLen)
						copy(b, v.Field(i).Bytes())
						_, err = buf.Write(b)
						if err != nil {
							return
						}
					} else {
						err = errors.New(fmt.Sprintf("Unknown slice type %s", v.Field(i).Kind()))
						return
					}
				case reflect.Struct:
					if t, ok := v.Field(i).Interface().(time.Time); ok {
						if tag, ok := tags[tagFormat]; ok {
							_, err = buf.WriteString(t.Format(tag))
							if err != nil {
								return
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
	res = buf.Bytes()
	return
}

func rightPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return retStr[:overallLen]
}
func leftPad2Len(s string, padStr string, overallLen int) string {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return retStr[(len(retStr) - overallLen):]
}
