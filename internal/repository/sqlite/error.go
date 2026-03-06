package sqliterepo

import "strings"

func isForeignKeyViolation(err error) bool {
	return strings.Contains(err.Error(), "FOREIGN KEY")
}
