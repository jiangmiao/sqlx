package sqlx

import "reflect"

func init() {
	tableNameMap = make(map[reflect.Type]string)
}
