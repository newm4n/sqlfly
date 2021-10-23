package sqlfly

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type DummyStruct struct {
	VisibleInt      int
	VisibleString   string
	VisibleBool     bool
	invisibleInt    int
	invisibleString string
	invisibleBool   bool
	VisibleTime     time.Time
}

func TestNewStructTable(t *testing.T) {
	_, err := NewStructTable(reflect.TypeOf(&DummyStruct{}), []string{"VisibleBool"})
	if err != nil {
		t.Error(err)
	}

	_, err = NewStructTable(reflect.TypeOf(&DummyStruct{}), []string{"invisibleString"})
	if err == nil {
		t.Error("should returned an error as fields are invisible")
	}

	_, err = NewStructTable(reflect.TypeOf(&DummyStruct{}), []string{"Visible"})
	if err == nil {
		t.Error("should returned an error as fields are non existent")
	}

	_, err = NewStructTable(reflect.TypeOf(&DummyStruct{}), []string{})
	if err != nil {
		t.Error(err)
	}

}

func TestColumnExist(t *testing.T) {
	st, err := NewStructTable(reflect.TypeOf(DummyStruct{}), []string{"VisibleString", "VisibleTime"})
	if err != nil {
		t.Error(err)
	}
	if st.columnExist("VisibleInt") != nil {
		t.Error("VisibleInt is visible")
	}
	if st.columnExist("VisibleString") != nil {
		t.Error("VisibleString is visible")
	}
	if st.columnExist("VisibleBool") != nil {
		t.Error("VisibleBool is visible")
	}
	if st.columnExist("invisibleInt") == nil {
		t.Error("invisibleInt is invisible")
	}
	if st.columnExist("invisibleString") == nil {
		t.Error("invisibleString is invisible")
	}
	if st.columnExist("invisibleBool") == nil {
		t.Error("invisibleBool is invisible")
	}
}

func TestStructTable_Insert(t *testing.T) {
	st, err := NewStructTable(reflect.TypeOf(DummyStruct{}), []string{"VisibleString", "VisibleTime"})
	if err != nil {
		t.Error(err)
	}

	if st.Count() != 0 {
		t.Errorf("Count error")
	}

	err = st.Insert(DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    11,
		invisibleString: "OneOne",
		invisibleBool:   true,
		VisibleTime:     time.Now(),
	})
	if err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond)

	if st.Count() != 1 {
		t.Errorf("Count error")
	}

	err = st.Insert(DummyStruct{
		VisibleInt:      2,
		VisibleString:   "Two",
		VisibleBool:     false,
		invisibleInt:    11,
		invisibleString: "OneOne",
		invisibleBool:   true,
		VisibleTime:     time.Now(),
	})
	if err != nil {
		t.Error(err)
	}

	if st.Count() != 2 {
		t.Errorf("Count error %d", st.Count())
	}
}

func TestStructTable_Insert_Duplicate(t *testing.T) {
	st, err := NewStructTable(reflect.TypeOf(DummyStruct{}), []string{"VisibleString", "VisibleTime"})
	if err != nil {
		t.Error(err)
	}

	if st.Count() != 0 {
		t.Errorf("Count error")
	}

	tNow := time.Now()

	err = st.Insert(DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    11,
		invisibleString: "OneOne",
		invisibleBool:   true,
		VisibleTime:     tNow,
	})
	if err != nil {
		t.Error(err)
	}

	time.Sleep(100 * time.Millisecond)

	if st.Count() != 1 {
		t.Errorf("Count error")
	}

	err = st.Insert(DummyStruct{
		VisibleInt:      2,
		VisibleString:   "Two",
		VisibleBool:     false,
		invisibleInt:    22,
		invisibleString: "TwoTwo",
		invisibleBool:   true,
		VisibleTime:     tNow,
	})
	if err == nil {
		t.Error("Duplicate error should be returned")
	}

	if st.Count() != 1 {
		t.Errorf("Count error %d", st.Count())
	}
}

func TestStructEquals(t *testing.T) {
	tNow := time.Now()
	if StructShallowEquals(DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    11,
		invisibleString: "OneOne",
		invisibleBool:   true,
		VisibleTime:     tNow,
	}, DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    12,
		invisibleString: "OneTwo",
		invisibleBool:   false,
		VisibleTime:     tNow,
	}) == false {
		t.Error("Struct is equal")
	}

	time.Sleep(100 * time.Millisecond)

	if StructShallowEquals(DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    11,
		invisibleString: "OneOne",
		invisibleBool:   true,
		VisibleTime:     tNow,
	}, DummyStruct{
		VisibleInt:      1,
		VisibleString:   "One",
		VisibleBool:     false,
		invisibleInt:    12,
		invisibleString: "OneTwo",
		invisibleBool:   false,
		VisibleTime:     time.Now(),
	}) == true {
		t.Error("Struct time is not equal")
	}
}

func GetSpell(i int) string {
	Spell := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
	byteArr := []byte(strconv.Itoa(i))
	var buff bytes.Buffer
	for _, i := range byteArr {
		buff.WriteString(Spell[i-48])
	}
	return buff.String()
}

func TestStructTable_Select(t *testing.T) {

	st, err := NewStructTable(reflect.TypeOf(DummyStruct{}), []string{"VisibleString"})
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 100; i++ {
		err := st.Insert(DummyStruct{
			VisibleInt:      i,
			VisibleString:   GetSpell(i),
			VisibleBool:     false,
			invisibleInt:    i + 100,
			invisibleString: GetSpell(i + 100),
			invisibleBool:   false,
			VisibleTime:     time.Now(),
		})
		if err != nil {
			t.Log("problem for i =", i)
		}
	}

	assert.Equal(t, 100, st.Count())

	ret, err := st.Select("VisibleString.contains(\"two\")", nil, 0, 0)
	assert.NoError(t, err, "there should be no error")
	assert.NotNilf(t, ret, "should be not nil")
	for idx, val := range ret {
		fmt.Printf("%d : %v\n", idx, val)
	}
}
