package fixedwidth

import (
	"testing"
	"bytes"
	"time"
	"strings"
)

func TestMarshalIntegerZeroPad(t *testing.T) {
	data := []byte("0001")
	src := struct {
		Number int `fixed:"len:4"`
	}{Number: 1}
	res, err := Marshal(&src)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Number decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalIntegerSpacePad(t *testing.T) {
	data := []byte("   1")
	src := struct {
		Number int `fixed:"len:4,pad: "`
	}{Number: 1}
	res, err := Marshal(src)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Number decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalIntegerZeroPadPtr(t *testing.T) {
	data := []byte("00010000")
	src := struct {
		Number1 *int `fixed:"len:4"`
		Number2 *int `fixed:"len:4"`
	}{}
	i := 1
	src.Number1 = &i
	res, err := Marshal(&src)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Number decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalIntegerSpacePadPtr(t *testing.T) {
	data := []byte("   1    ")
	src := struct {
		Number1 *int `fixed:"len:4,pad: "`
		Number2 *int `fixed:"len:4,pad: "`
	}{}
	i := 1
	src.Number1 = &i
	res, err := Marshal(src)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Number decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalIntegerHEX(t *testing.T) {
	data := []byte("00FF")
	dest := struct {
		Number int `fixed:"len:4,base:16"`
	}{Number: 255}
	res, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Number decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalString(t *testing.T) {
	data := []byte("TEST    ")
	dest := struct {
		String string `fixed:"len:8"`
	}{String: "TEST"}
	res, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("String decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalSkipNontagged(t *testing.T) {
	data := []byte("TEST    ")
	dest := struct {
		String1 string `fixed:"len:8"`
		String2 string
	}{String1: "TEST", String2: "STUFF"}
	res, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("String decoded incorrectly expected:", data, "got:", res)
	}
}
func TestMarshalBytes(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	dest := struct {
		Bytes []byte `fixed:"len:4"`
	}{Bytes: []byte{0xDE, 0xAD, 0xBE, 0xEF}}
	res, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(res, data) != 0 {
		t.Error("Bytes decoded incorrectly expected:", data, "got:", res)
	}
}

func TestMarshalTime(t *testing.T) {
	dest := struct {
		Date time.Time `fixed:"len:8,format:01022006"`
	}{Date: time.Now()}
	res, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	r := string(res)
	data := dest.Date.Format("01022006")
	if strings.Compare(data, r) != 0 {
		t.Error("Time incorrectly expected: '" + data + "' got: '" + r + "'")
	}
}

func TestMultiMarshal(t *testing.T) {
	data := []byte("11161990123Hello")

	dest := struct {
		Date   time.Time `fixed:"len:8,format:01022006"`
		Number int `fixed:"len:3"`
		String string `fixed:"len:5"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}

	b, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(b, data) != 0 {
		t.Error("MultiStruct incorrectly expected:", data, "got:", b)
	}
}

func TestMarshalCustom(t *testing.T) {
	dest := struct {
		Date1 FixedDate `fixed:"len:8"`
		Date2 *FixedDate `fixed:"len:8"`
	}{}
	dest.Date1.Time = time.Now()
	data, err := Marshal(dest)
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(dest.Date1.Format("01022006")+"        ", string(data)) != 0 {
		t.Error("Date decoded incorrectly, expected: '" + string(data) + "' got: '" + dest.Date1.Format("01022006") + "'")
	}
}
