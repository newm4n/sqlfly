package sqlfly

const (
	TypeString ColumnType = iota
	TypeBool
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeDateTime

	LogicOpEQ LogicalOperator = iota
	LogicOpNEQ
	LogicOpLT
	LogicOpLTE
	LogicOpGT
	LogicOpGTE
)

type ColumnType int
type LogicalOperator int

type ColumnMeta struct {
	Name string
	Sequence int
	Unique bool
	Nullable bool
	Type ColumnType
}

type TableMeta struct {
	Name string
	ColumnMetas *ColumnMeta
}

type KeyValue struct {
	Key string
	Value interface{}
}

type Filter struct {
	Key string
	Operator LogicalOperator
	Value interface{}
}

type Table struct {
	Meta *TableMeta
}

type Row struct {
	Data []*KeyValue
}

type Rows struct {
	Data []*Row
}

type Driver interface {
	Select(from string, fields []string) (Rows *Rows, err error)
	Insert(into string, fields []string, values []interface{}) (int, error)
	Update(table string,set []*KeyValue, filters []*Filter) (int, error)
	Delete(from string, filters []*Filter) (int, error)
}