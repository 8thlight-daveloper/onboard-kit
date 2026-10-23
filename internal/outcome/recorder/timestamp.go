package recorder

import (
	"regexp"
	"strconv"
)

func validTimestamp(value string) bool {
	matched, _ := regexp.MatchString(`^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]{1,9})?(Z|\+00:00)$`, value)
	if !matched {
		return false
	}
	year, _ := strconv.Atoi(value[:4])
	if year == 0 {
		return false
	}
	month, _ := strconv.Atoi(value[5:7])
	day, _ := strconv.Atoi(value[8:10])
	days := [...]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		days[1] = 29
	}
	return day <= days[month-1]
}
