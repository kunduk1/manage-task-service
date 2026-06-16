package db

import (
	stderrors "errors"

	driver "github.com/go-sql-driver/mysql"
)

// mysqlDuplicateEntry — код ошибки MySQL для нарушения UNIQUE/PRIMARY-ограничения.
const mysqlDuplicateEntry = 1062

// IsDuplicateEntry сообщает, вызвана ли ошибка нарушением UNIQUE/PRIMARY-ограничения MySQL.
func IsDuplicateEntry(err error) bool {
	var mysqlErr *driver.MySQLError
	return stderrors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDuplicateEntry
}
