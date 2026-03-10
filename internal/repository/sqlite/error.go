package sqliterepo

import "strings"

func isForeignKeyViolation(err error) bool {
	return strings.Contains(err.Error(), "FOREIGN KEY")
}

func isUniqueConstraintViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE")
}
