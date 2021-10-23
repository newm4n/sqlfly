package sqlfly

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"log"
	"reflect"
	"time"
)

const (
	OpEqual Operator = iota
	OpNotEqual
	OpLessThan
	OpLessThanEqual
	OpGreaterThan
	OpGreaterThanEqual

	OrderASC Ordering = iota
	OrderDESC

	BaseKindInt BaseKind = iota
	BaseKindUint
	BaseKindFloat
	BaseKindBool
	BaseKindString
	BaseKindTime
	BaseKindOther
)

type BaseKind int
type Operator int
type Ordering int

type Order struct {
	Key      string
	Ordering Ordering
}

type KeyValue struct {
	Key   string
	Value reflect.Value
}

type KeyValueCompare struct {
	KeyValue
	Op Operator
}

func NewStructTable(metaType reflect.Type, uniques []string) (*StructTable, error) {
	ret := &StructTable{
		MetaType: metaType,
		Uniques:  uniques,
		dataSet:  make([]interface{}, 0),
	}

	if metaType.Kind() == reflect.Ptr {
		ret.MetaType = metaType.Elem()
	}
	if err := ret.validate(); err != nil {
		return nil, err
	}
	return ret, nil
}

type StructTable struct {
	MetaType reflect.Type
	Uniques  []string
	dataSet  []interface{}
}

func (st *StructTable) validate() error {
	// check if MetaType is a struct
	if st.MetaType.Kind() != reflect.Struct {
		return ErrMetaNotAStruct
	}

	// check for keys existance.
	if err := st.columnsExist(st.Uniques); err != nil {
		return ErrUniquesColumnNameNotExist
	}

	return nil
}

func (st *StructTable) columnsExist(columnNames []string) error {
	if len(columnNames) > 0 {
		for _, k := range columnNames {
			if len(k) == 0 {
				return ErrColumnNameEmpty
			}
			if err := st.columnExist(k); err != nil {
				return err
			}
		}
	}
	return nil
}

func (st *StructTable) columnExist(columnName string) error {
	if _, exist := st.MetaType.FieldByName(columnName); exist {
		byteArr := []byte(columnName)
		if byteArr[0] >= 65 && byteArr[0] <= 90 {
			return nil
		}
	}
	return ErrColumnNameNotExist
}

func (st *StructTable) Insert(data interface{}) error {
	dataType := reflect.TypeOf(data)
	if dataType.Kind() != reflect.Struct {
		return ErrCanNotInsertNoStructType
	}
	if dataType.String() != st.MetaType.String() {
		return ErrIncompatibleStruct
	}

	dataValue := reflect.ValueOf(data)

	for i := 0; i < dataType.NumField(); i++ {
		dataFieldTyp := dataType.Field(i)
		if err := st.columnExist(dataFieldTyp.Name); err != nil {
			continue
		}
		// check for nulls.
		switch dataFieldTyp.Type.Kind() {
		case reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
			continue
		case reflect.Struct:
			if dataFieldTyp.Type.String() != "time.Time" {
				continue
			}
		}
		dataFieldValue := dataValue.Field(i)

		// check for uniqueness
		if Contains(st.Uniques, dataFieldTyp.Name) {
			for _, tableRow := range st.dataSet {
				tableRowValue := reflect.ValueOf(tableRow)
				tableFieldValue := tableRowValue.Field(i)
				switch dataFieldTyp.Type.Kind() {
				case reflect.Bool:
					if tableFieldValue.Bool() == dataFieldValue.Bool() {
						return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if tableFieldValue.Int() == dataFieldValue.Int() {
						return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					if tableFieldValue.Uint() == dataFieldValue.Uint() {
						return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
					}
				case reflect.Float32, reflect.Float64:
					if tableFieldValue.Float() == dataFieldValue.Float() {
						return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
					}
				case reflect.String:
					if tableFieldValue.String() == dataFieldValue.String() {
						return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
					}
				case reflect.Struct:
					if dataFieldTyp.Type.String() == "time.Time" {
						dfTime := tableFieldValue.Interface().(time.Time)
						fTime := dataFieldValue.Interface().(time.Time)
						if dfTime == fTime {
							return fmt.Errorf("%w :  %s", ErrUniqueConstraintViolation, dataFieldTyp.Name)
						}
					}
				}
			}
		}
	}
	st.dataSet = append(st.dataSet, data)
	return nil
}

func (st *StructTable) filter(where string) ([]int, error) {
	declarations := make([]*exprpb.Decl, 0)
	fieldNames := make([]string, 0)
	for i := 0; i < st.MetaType.NumField(); i++ {
		rowFieldType := st.MetaType.Field(i)
		byteArr := []byte(rowFieldType.Name)
		head := byteArr[0]
		if head < 65 || head > 90 {
			continue
		}
		add := false
		switch rowFieldType.Type.Kind() {
		case reflect.Int, reflect.String, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float64, reflect.Float32, reflect.Bool:
			add = true
		case reflect.Struct:
			if rowFieldType.Type.String() == "time.Time" {
				add = true
			}
		}
		if add {
			fieldNames = append(fieldNames, rowFieldType.Name)
		}
	}
	for _, name := range fieldNames {
		rFieldValue, _ := st.MetaType.FieldByName(name)
		var typ *exprpb.Type
		switch GetBaseKindOfType(rFieldValue.Type) {
		case BaseKindFloat:
			typ = decls.Double
		case BaseKindBool:
			typ = decls.Bool
		case BaseKindUint:
			typ = decls.Uint
		case BaseKindInt:
			typ = decls.Int
		case BaseKindString:
			typ = decls.String
		case BaseKindTime:
			typ = decls.Timestamp
		}
		declarations = append(declarations, decls.NewVar(name, typ))
	}
	env, err := cel.NewEnv(
		cel.Declarations(declarations...))
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(where)
	if issues != nil && issues.Err() != nil {
		log.Fatalf("type-check error: %s", issues.Err())
		return nil, err
	}
	prg, err := env.Program(ast)
	if err != nil {
		log.Fatalf("program construction error: %s", err)
		return nil, err
	}
	ret := make([]int, 0)
	for index, row := range st.dataSet {
		result, _, err := prg.Eval(ToMap(row))
		if err != nil {
			return nil, err
		}
		if reflect.ValueOf(result.Value()).Kind() != reflect.Bool {
			return nil, fmt.Errorf("%w : expression do not yield boolean result [%s]", ErrEvaluationError, where)
		}
		if result.Value().(bool) {
			ret = append(ret, index)
		}
	}
	return ret, nil
}

func (st *StructTable) Update(set []KeyValue, where string) (int, error) {
	return 0, nil
}

func (st *StructTable) Delete(where string) (int, error) {
	return 0, nil
}

func (st *StructTable) Select(where string, orderBy []Order, offset, len int) ([]interface{}, error) {
	recIdxs, err := st.filter(where)
	if err != nil {
		return nil, err
	}
	ret := make([]interface{}, 0)
	for _, idx := range recIdxs {
		ret = append(ret, st.dataSet[idx])
	}
	// TODO, implement the offset, len and ordering
	return ret, nil
}

func (st *StructTable) Count() int {
	return len(st.dataSet)
}

// Contains check the existence of a string in string array
func Contains(arr []string, s string) bool {
	for _, ss := range arr {
		if s == ss {
			return true
		}
	}
	return false
}

func StructShallowEquals(one, two interface{}) bool {
	oneType := reflect.TypeOf(one)
	oneVal := reflect.ValueOf(one)

	twoType := reflect.TypeOf(two)
	twoVal := reflect.ValueOf(two)

	if oneType.Kind() != reflect.Struct || oneType.Kind() != reflect.Struct || oneType.String() != twoType.String() || oneType.NumField() != twoType.NumField() {
		return false
	}

	for i := 0; i < oneType.NumField(); i++ {
		oneFieldValue := oneType.Field(i)
		twoFieldValue := twoType.Field(i)
		if oneFieldValue.Name != twoFieldValue.Name {
			return false
		}
		if oneFieldValue.Type.Kind() != twoFieldValue.Type.Kind() {
			return false
		}
		byteArr := []byte(oneFieldValue.Name)
		head := byteArr[0]
		if head < 65 || head > 90 {
			continue
		}

		oneFieldDataValue := oneVal.Field(i)
		twoFieldDataValue := twoVal.Field(i)

		switch oneFieldValue.Type.Kind() {
		case reflect.Bool:
			if oneFieldDataValue.Bool() != twoFieldDataValue.Bool() {
				return false
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if oneFieldDataValue.Int() != twoFieldDataValue.Int() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if oneFieldDataValue.Uint() != twoFieldDataValue.Uint() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if oneFieldDataValue.Float() != twoFieldDataValue.Float() {
				return false
			}
		case reflect.String:
			if oneFieldDataValue.String() != twoFieldDataValue.String() {
				return false
			}
		case reflect.Struct:
			if oneFieldValue.Type.String() == "time.Time" {
				dfTime := oneFieldDataValue.Interface().(time.Time)
				fTime := twoFieldDataValue.Interface().(time.Time)
				if dfTime != fTime {
					return false
				}
			}
		}
	}
	return true
}

func GetBaseKindOfType(typ reflect.Type) BaseKind {
	switch typ.Kind() {
	case reflect.Bool:
		return BaseKindBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return BaseKindInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return BaseKindUint
	case reflect.Float32, reflect.Float64:
		return BaseKindFloat
	case reflect.String:
		return BaseKindString
	case reflect.Struct:
		if typ.String() == "time.Time" {
			return BaseKindTime
		}
		return BaseKindOther
	default:
		return BaseKindOther
	}
}

func GetBaseKind(value reflect.Value) BaseKind {
	switch value.Type().Kind() {
	case reflect.Bool:
		return BaseKindBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return BaseKindInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return BaseKindUint
	case reflect.Float32, reflect.Float64:
		return BaseKindFloat
	case reflect.String:
		return BaseKindString
	case reflect.Struct:
		if value.Type().String() == "time.Time" {
			return BaseKindTime
		}
		return BaseKindOther
	default:
		return BaseKindOther
	}
}

func ToMap(strck interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	sType := reflect.TypeOf(strck)
	sValue := reflect.ValueOf(strck)

	for i := 0; i < sType.NumField(); i++ {
		typeField := sType.Field(i)
		byteArr := []byte(typeField.Name)
		head := byteArr[0]
		if head < 65 || head > 90 || GetBaseKindOfType(typeField.Type) == BaseKindOther {
			continue
		}
		valueField := sValue.Field(i)
		switch GetBaseKindOfType(typeField.Type) {
		case BaseKindInt:
			ret[typeField.Name] = valueField.Int()
		case BaseKindUint:
			ret[typeField.Name] = valueField.Uint()
		case BaseKindFloat:
			ret[typeField.Name] = valueField.Float()
		case BaseKindBool:
			ret[typeField.Name] = valueField.Bool()
		case BaseKindString:
			ret[typeField.Name] = valueField.String()
		case BaseKindTime:
			ret[typeField.Name] = valueField.Interface().(time.Time)
		}
	}
	return ret
}
