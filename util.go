package fixedwidth

import (
	"strings"
	"reflect"
	"strconv"
)

type fixedTags struct {
	Len    int
	Pad    string
	Format string
	Base   int
	Align  string
}

func parseTags(field reflect.StructField, kind reflect.Kind) (f *fixedTags, err error) {
	tags := make(map[string]string)
	tag := field.Tag.Get(tagName)
	if tag == "" {
		return
	}
	var rawTags = strings.Split(tag, ",")

	for _, rt := range rawTags {
		x := strings.Split(rt, ":")
		tags[x[0]] = x[1]
	}
	f = new(fixedTags)
	if f.Len, err = strconv.Atoi(tags[tagLen]); err != nil {
		return
	}
	f.Base = 10
	if b, ok := tags[tagBase]; ok {
		if f.Base, err = strconv.Atoi(b); err != nil {
			return
		}
	}
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f.Pad = defaultPadInt
		f.Align = alignRight
	default:
		f.Align = alignLeft
		f.Pad = defaultPadString
	}
	if t, ok := tags[tagPad]; ok {
		f.Pad = t
	}
	f.Format = tags[tagFormat]
	if t, ok := tags[tagAlign]; ok {
		f.Align = t
	}
	return
}

// unused for now but will probably add an "align" tag
func rightPad2Len(s string, padStr string, overallLen int) []byte {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = s + strings.Repeat(padStr, padCountInt)
	return []byte(retStr[:overallLen])
}

func leftPad2Len(s string, padStr string, overallLen int) []byte {
	var padCountInt int
	padCountInt = 1 + ((overallLen - len(padStr)) / len(padStr))
	var retStr = strings.Repeat(padStr, padCountInt) + s
	return []byte(retStr[(len(retStr) - overallLen):])
}

func alignAndPad2Len(align string, s string, padStr string, overallLen int) []byte {
	if align == alignLeft {
		return rightPad2Len(s, padStr, overallLen)
	} else if align == alignRight {
		return leftPad2Len(s, padStr, overallLen)
	}
	return []byte{}
}

// unused for now but will probably need for pointers and stoof
func initializeStruct(t reflect.Type, v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)
		switch ft.Type.Kind() {
		case reflect.Map:
			f.Set(reflect.MakeMap(ft.Type))
		case reflect.Slice:
			f.Set(reflect.MakeSlice(ft.Type, 0, 0))
		case reflect.Chan:
			f.Set(reflect.MakeChan(ft.Type, 0))
		case reflect.Struct:
			initializeStruct(ft.Type, f)
		case reflect.Ptr:
			fv := reflect.New(ft.Type.Elem())
			initializeStruct(ft.Type.Elem(), fv.Elem())
			f.Set(fv)
		default:
		}
	}
}
