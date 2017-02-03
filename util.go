package fixedwidth

import (
	"strings"
	"reflect"
)

func splitTags(field reflect.StructField) (tags map[string]string) {
	tags = make(map[string]string)
	tag := field.Tag.Get(tagName)
	if tag == "" {
		return
	}
	var rawTags = strings.Split(tag, ",")

	for _, rt := range rawTags {
		x := strings.Split(rt, ":")
		tags[x[0]] = x[1]
	}
	return
}
