package sqlfly

import "fmt"

var (
	// ErrMetaNotAStruct error returned if the specified meta is not a struct
	ErrMetaNotAStruct = fmt.Errorf("the table meta data is not a struct")

	// ErrInvalidColumnName error returned if the specified column name is not valid
	ErrInvalidColumnName = fmt.Errorf("invalid column name")

	// ErrColumnNameEmpty error returned if any of the specified column name is an empty string
	ErrColumnNameEmpty = fmt.Errorf("%w : specified column name is empty", ErrInvalidColumnName)

	// ErrColumnNameNotExist error returned if any of the specified column name is not matching any field in the meta type
	ErrColumnNameNotExist = fmt.Errorf("%w : specified column name not exist in struct", ErrInvalidColumnName)

	// ErrUniquesColumnNameNotExist error returned if any othe specified unique column name is not matching any field in the meta type
	ErrUniquesColumnNameNotExist = fmt.Errorf("%w : unique column", ErrColumnNameNotExist)

	// ErrNullsColumnNameNotExist error returned if any of the specified nullable column name is not matching any field of the meta type
	ErrNullsColumnNameNotExist = fmt.Errorf("%w : null column", ErrColumnNameNotExist)

	// ErrInsertError errors returned when theres a problem during inserting a data into StructTable
	ErrInsertError = fmt.Errorf("problem when inserting data")

	// ErrCanNotInsertNoStructType errors returned when the data to insert is not a struct
	ErrCanNotInsertNoStructType = fmt.Errorf("%w : can not insert non struct", ErrInsertError)

	// ErrCanNotInsertNonNativeType errors returned when the data to insert is not a struct
	ErrCanNotInsertNonNativeType = fmt.Errorf("%w : can not insert non native field", ErrInsertError)

	// ErrIncompatibleStruct errors returned when data to insert is not compatible with the meta struct
	ErrIncompatibleStruct = fmt.Errorf("%w : incompatible struct to insert", ErrInsertError)

	// ErrUniqueConstraintViolation error returned when a inserted field value is duplicated violating unique constraint
	ErrUniqueConstraintViolation = fmt.Errorf("%w : unique field constraint violation", ErrInsertError)

	// ErrEvaluationError errors returned when cel-go evaluation yield error
	ErrEvaluationError = fmt.Errorf("problem when evaluating data")
)
