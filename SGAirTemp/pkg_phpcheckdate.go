package SGAirTemp

//Based on: https://github.com/openset/php2go/blob/master/php/checkdate.go
//Don't want to complicate the integration, so copy to another file as this package, but credit should goes to him

import "strconv"

//Checkdate - Validate a Gregorian date.
func Checkdate(mo, dt, yr string) bool {
	var month, day, year uint64
	month, _ = strconv.ParseUint(mo, 10, 32)
	day, _ = strconv.ParseUint(dt, 10, 32)
	year, _ = strconv.ParseUint(yr, 10, 32)
	//check month
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		if day > 31 {
			return false
		}
	case 4, 6, 9, 11:
		if day > 30 {
			return false
		}
	case 2:
		if checkIfLeapYear(year) {
			if day > 29 {
				return false
			}
		} else {
			if day > 28 {
				return false
			}
		}
	default:
		return false
	}

	//check day
	if day < 1 {
		return false
	}

	//check year
	if year < 1 || year > 32767 {
		return false
	}

	return true
}

func checkIfLeapYear(year uint64) bool {

	if year%100 == 0 {
		if year%400 == 0 {
			return true
		}
		return false
	}
	if year%4 == 0 {
		return true
	}

	return false
}
