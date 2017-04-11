package fixedwidth

import (
	"testing"
	"bytes"
	"encoding/hex"
	"time"
	"strings"
)

func TestUnmarshalIntegerZeroPad(t *testing.T) {
	data := []byte("0001")
	dest := struct {
		Number int `fixed:"len:4"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Number != 1 {
		t.Error("Number decoded incorrectly expected: 1 got:", dest.Number)
	}
}
func TestUnmarshalIntegerSpacePad(t *testing.T) {
	data := []byte("   1    ")
	dest := struct {
		Number1 int `fixed:"len:4,pad: "`
		Number2 int `fixed:"len:4,pad: "`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Number1 != 1 {
		t.Error("Number decoded incorrectly expected: 1 got:", dest.Number1)
	}
	if dest.Number2 != 0 {
		t.Error("Number decoded incorrectly expected: 0 got:", dest.Number2)
	}
}
func TestUnmarshalIntegerZeroPadPtr(t *testing.T) {
	data := []byte("00010000")
	dest := struct {
		Number1 *int `fixed:"len:4"`
		Number2 *int `fixed:"len:4"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Number1 == nil || *dest.Number1 != 1 {
		t.Error("Number decoded incorrectly expected: 1 got:", dest.Number1)
	}
	if dest.Number1 == nil || *dest.Number2 != 0 {
		t.Error("Number decoded incorrectly expected: 0 got:", dest.Number1)
	}
}
func TestUnmarshalIntegerSpacePadPtr(t *testing.T) {
	data := []byte("   1    ")
	dest := struct {
		Number1 *int `fixed:"len:4,pad: "`
		Number2 *int `fixed:"len:4,pad: "`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Number1 == nil || *dest.Number1 != 1 {
		t.Error("Number decoded incorrectly expected: 1 got:", dest.Number1)
	}
	if dest.Number2 != nil {
		t.Error("Number decoded incorrectly expected: nil")
	}
}
func TestUnmarshalIntegerHEX(t *testing.T) {
	data := []byte("00FF")
	dest := struct {
		Number int `fixed:"len:4,base:16"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Number != 255 {
		t.Error("Number decoded incorrectly expected: 255 got:", dest.Number)
	}
}
func TestUnmarshalString(t *testing.T) {
	data := []byte("TEST    ")
	dest := struct {
		String string `fixed:"len:8"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}

	if dest.String != "TEST" {
		t.Error("String decoded incorrectly expected: 'TEST' got:'" + dest.String + "'")
	}
}
func TestUnmarshalBytes(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	dest := struct {
		Bytes []byte `fixed:"len:4"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(dest.Bytes, []byte{0xDE, 0xAD, 0xBE, 0xEF}) != 0 {
		t.Error("Bytes decoded incorrectly expected: DEADBEEF got:", hex.EncodeToString(dest.Bytes))
	}
}

func TestUnmarshalTime(t *testing.T) {
	data := []byte("11161990")

	dest := struct {
		Date time.Time `fixed:"len:8,format:01022006"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(dest.Date.Format("01022006"), string(data)) != 0 {
		t.Error("Date decoded incorrectly")
	}
}
func TestUnmarshalTimePtr(t *testing.T) {
	data := []byte("11161990        ")

	dest := struct {
		Date1 *time.Time `fixed:"len:8,format:01022006"`
		Date2 *time.Time `fixed:"len:8,format:01022006"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(dest.Date1.Format("01022006"), "11161990") != 0 {
		t.Error("Date decoded incorrectly got", dest.Date1.Format("01022006"))
	}
	if dest.Date2 != nil {
		t.Error("Date decoded incorrectly expected nil")
	}
}

type FixedDate struct {
	time.Time
}

func (f *FixedDate) UnmarshalFixed(data []byte) (err error) {
	f.Time, err = time.Parse("01022006", string(data))
	return
}

func TestUnmarshalCustom(t *testing.T) {
	data := []byte("11161990")

	dest := struct {
		Date *FixedDate `fixed:"len:8"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.Date == nil {
		t.Error("Date decoded incorrectly, got nil")
	} else if strings.Compare(dest.Date.Format("01022006"), string(data)) != 0 {
		t.Error("Date decoded incorrectly, expected: '" + string(data) + "' got: '" + dest.Date.Format("01022006") + "'")
	}
}

func TestMultiUnmarshal(t *testing.T) {
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
}

func TestUnmarshalStringPtr(t *testing.T) {
	data := []byte("TEST            ")
	dest := struct {
		String1 *string `fixed:"len:8"`
		String2 *string `fixed:"len:8"`
	}{}
	err := Unmarshal(data, &dest)
	if err != nil {
		t.Error(err)
	}
	if dest.String1 == nil {
		t.Error("String1 decoded incorrectly expected: TEST, got nil")
	} else {
		if *dest.String1 != "TEST" {
			t.Error("String1 decoded incorrectly expected: TEST")
		}
	}
	if dest.String2 != nil {
		t.Error("String2 decoded incorrectly expected: nil")
	}
}
