package fixedwidth

import (
	"reflect"
	"bytes"
	"errors"
	"io"
	"strconv"
	"fmt"
	"time"
)

type Marshaler interface {
	MarshalFixed() ([]byte, error)
}

func Marshal(in interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	err := marshalRecursive(&buf, nil, reflect.ValueOf(in))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
func marshalRecursive(w io.Writer, field *reflect.StructField, val reflect.Value) (err error) {
	var tag *fixedTags
	if field != nil {
		tag, err = parseTags(*field, val.Kind())
		if err != nil {
			return
		}
	}

	switch val.Kind() {
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		unwrapped := val.Elem()

		// Check if the pointer is nil
		if unwrapped.IsValid() {
			return marshalRecursive(w, field, unwrapped)
		} else {
			newInst := reflect.New(val.Type().Elem())
			tag, err = parseTags(*field, newInst.Elem().Kind())
			if err != nil {
				return
			}
			strInt := alignAndPad2Len(tag.Align,"", tag.Pad, tag.Len)
			_, err = w.Write(strInt)
			return
		}
		return marshalRecursive(w, field, unwrapped)
	case reflect.Interface:
		unwrapped := val.Elem()
		return marshalRecursive(w, field, unwrapped)
	case reflect.Struct:
		//custom marshaler
		if _, ok := val.Interface().(Marshaler); ok {
			var b []byte
			b, err = val.Interface().(Marshaler).MarshalFixed()
			if err != nil {
				return
			}
			_, err = w.Write(b)
			return
		}

		// struct type exceptions
		if t, ok := val.Interface().(time.Time); ok {
			if tag.Format != "" {
				w.Write([]byte(t.Format(tag.Format)))
				return
			} else {
				err = errors.New("no date format specified")
				return
			}
		}

		// else walk the fields
		tipe := reflect.TypeOf(val.Interface())
		for i := 0; i < val.NumField(); i += 1 {
			field := tipe.Field(i)
			tag := field.Tag.Get(tagName)
			if tag == "" {
				continue
			}
			if err = marshalRecursive(w, &field, val.Field(i)); err != nil {
				return
			}

		}
	case reflect.String:
		strInt := alignAndPad2Len(tag.Align,val.String(), tag.Pad, tag.Len)
		_, err = w.Write(strInt)
		return
	case reflect.Int, reflect.Int32, reflect.Int64:
		strInt := alignAndPad2Len(tag.Align,strconv.FormatInt(val.Int(), tag.Base), tag.Pad, tag.Len)

		// always do upper case for hex and stuff
		if tag.Base != 10 {
			strInt = bytes.ToUpper(strInt)
		}
		_, err = w.Write(strInt)
		return
	case reflect.Slice:
		if _, ok := val.Interface().([]byte); ok {
			b := bytes.Repeat([]byte(tag.Pad), tag.Len)
			copy(b, val.Bytes())
			_, err = w.Write(b)
			return
		} else {
			err = errors.New(fmt.Sprintf("Unknown slice type %s", val.Kind()))
			return
		}

	default:
		err = errors.New("Unknown type: " + val.Kind().String())
		return
	}
	return
}

//
//func Marshal1(in interface{}) (res []byte, err error) {
//	t := reflect.TypeOf(in)
//	v := reflect.ValueOf(in)
//	if reflect.TypeOf(in).Kind() == reflect.Ptr {
//		t = t.Elem()
//		v = v.Elem()
//	}
//
//	buf := bytes.Buffer{}
//
//	for i := 0; i < t.NumField(); i++ {
//		field := t.Field(i)
//		tag := field.Tag.Get(tagName)
//		if tag == "" {
//			continue
//		}
//		var fieldLen int
//		tags := splitTags(field)
//		fieldLen, err = strconv.Atoi(tags[tagLen])
//		if err != nil {
//			err = errors.New(fmt.Sprintf("fieldLen %s, %s", field.Name, err.Error()))
//			return
//		}
//		if len(tag) > 0 {
//			if _, ok := v.Field(i).Interface().(Marshaler); ok {
//				if v.Field(i).IsNil() {
//					v.Field(i).Set(reflect.New(v.Field(i).Type().Elem()))
//				}
//				var b []byte
//				b, err = v.Field(i).Interface().(Marshaler).MarshalFixed()
//				if err != nil {
//					return
//				}
//				_, err = buf.Write(b)
//				if err != nil {
//					return
//				}
//			} else {
//				switch v.Field(i).Kind() {
//				case reflect.Int, reflect.Int32, reflect.Int64:
//					var pad = defaultPadInt
//					if p, ok := tags[tagPad]; ok {
//						pad = p
//					}
//					var base = 10
//					if _, ok := tags[tagBase]; ok {
//						base, err = strconv.Atoi(tags[tagBase])
//						if err != nil {
//							err = errors.New("Invalid integer base " + tags[tagBase])
//							return
//						}
//					}
//
//					strInt := alignAndPad2Len(tag.Align,strconv.FormatInt(v.Field(i).Int(), base), pad, fieldLen)
//
//					// always do upper case for hex and stuff
//					if base != 10 {
//						strInt = strings.ToUpper(strInt)
//					}
//
//					_, err = buf.WriteString(strInt)
//					if err != nil {
//						return
//					}
//				case reflect.String:
//					var pad = defaultPadString
//					if p, ok := tags[tagPad]; ok {
//						pad = p
//					}
//					strInt := alignAndPad2Len(tag.Align,v.Field(i).String(), pad, fieldLen)
//					_, err = buf.WriteString(strInt)
//					if err != nil {
//						return
//					}
//				case reflect.Slice:
//					if _, ok := v.Field(i).Interface().([]byte); ok {
//						pad := []byte{0x0}
//						if p, ok := tags[tagPad]; ok {
//							pad = []byte(p)
//						}
//						b := bytes.Repeat(pad, fieldLen)
//						copy(b, v.Field(i).Bytes())
//						_, err = buf.Write(b)
//						if err != nil {
//							return
//						}
//					} else {
//						err = errors.New(fmt.Sprintf("Unknown slice type %s", v.Field(i).Kind()))
//						return
//					}
//				case reflect.Struct:
//					if t, ok := v.Field(i).Interface().(time.Time); ok {
//						if tag, ok := tags[tagFormat]; ok {
//							_, err = buf.WriteString(t.Format(tag))
//							if err != nil {
//								return
//							}
//						} else {
//							err = errors.New("no date format specified")
//							return
//						}
//					} else {
//						err = errors.New(fmt.Sprintf("Unknown struct %s", v.Field(i).Kind()))
//						return
//					}
//				default:
//					err = errors.New(fmt.Sprintf("Unknown kind %s", v.Field(i).Kind()))
//					return
//				}
//			}
//		}
//	}
//	res = buf.Bytes()
//	return
//}
