package SGAirTemp

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//CheckInputDate - a function to validated the input date
func CheckInputDate(StrDate string) (validatedDate string) {
	currentTime := time.Now()
	curDateTimeStr := currentTime.Format("2006-01-02T15:04:05")
	curDateTimeArr := strings.Split(curDateTimeStr, "T")
	curDateStr := curDateTimeArr[0]

	//Remove the carriage returns for Linux/Windows - assuming user input manually
	validatedDate = strings.TrimRight(StrDate, "\r\n")
	validatedDate = strings.TrimRight(validatedDate, "\n")
	//Trim the spaces
	validatedDate = strings.TrimSpace(validatedDate)

	//Date is not inputted by the user, we use current date
	if len(validatedDate) == 0 {
		validatedDate = curDateStr
	}

	//Regex Format for YYYY-MM-DD
	datePattern := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

	if datePattern.MatchString(validatedDate) == false {
		fmt.Printf("The inputed Data '%s' is not match with YYYY-MM-DD format.\n", validatedDate)
		os.Exit(1)
	}

	//Split the DateVal, so we can get the individual year, month and date.
	arrDate := strings.Split(validatedDate, "-")

	//Ensure the date is greater than 1970, for Unix time.
	testYear, _ := strconv.Atoi(arrDate[0])
	if testYear < 1970 {
		fmt.Printf("Please enter the year above or equals 1970.\n")
		os.Exit(1)
	}

	//Check if the date inputted is a valid date.
	if Checkdate(arrDate[1], arrDate[2], arrDate[0]) == false {
		fmt.Printf("The inputed Data '%s' not a valid date. Please check again.\n", validatedDate)
		os.Exit(1)
	}

	//Create the Inputted Date/Time object, the created object will without UTC/Timezone
	InputDate, _ := time.Parse(strStandardFormat, validatedDate)

	//Unix Value for Inputted DateTime.
	InputDateUnix := InputDate.Unix()

	//Create the Date object of the EarliestDataAvail
	EarliestDate, _ := time.Parse(strStandardFormat, EarliestDataAvail)

	//Unix Value for Earliest Date.
	EarliestDateUnix := EarliestDate.Unix()

	if InputDateUnix < EarliestDateUnix {
		fmt.Printf("The earliest Data we had for this API is: '%s'\n Please refer to https://data.gov.sg/dataset/realtime-weather-readings under the 'Coverage'", EarliestDataAvail)
		os.Exit(1)
	}

	return validatedDate
}

//CheckInputTime - a function to validated the input time
func CheckInputTime(StrTime string) (validatedTime string) {
	currentTime := time.Now()
	curDateTimeStr := currentTime.Format("2006-01-02T15:04:05")
	curDateTimeArr := strings.Split(curDateTimeStr, "T")
	curTimeStr := string([]rune(curDateTimeArr[1])[0:5])

	//Remove the carriage returns for Linux/Windows
	validatedTime = strings.TrimRight(StrTime, "\r\n")
	validatedTime = strings.TrimRight(validatedTime, "\n")
	//Trim the spaces
	validatedTime = strings.TrimSpace(validatedTime)

	//Time is not inputted by the user, we use current time
	if len(validatedTime) == 0 {
		validatedTime = curTimeStr
	}

	arrTime := strings.Split(validatedTime, ":")
	intHour, _ := strconv.Atoi(arrTime[0])
	intMin, _ := strconv.Atoi(arrTime[1])
	if intHour < 0 || intHour > 23 || intMin < 0 || intMin > 59 {
		fmt.Printf("The inputed Data '%s' not a valid time. Please check again.\n", validatedTime)
		os.Exit(1)
	}

	return validatedTime
}

//ValidateDateTimeLessThanNow - a function to validated the input date
func ValidateDateTimeLessThanNow(ValDate, ValTime string) (validatedResult bool) {
	//Standardized format
	strStandardFormat := "2006-01-02T15:04:05"

	//Get Current Date/Time
	currentTime := time.Now()

	//Get Formatted Current Date Time
	curDateTimeStr := currentTime.Format(strStandardFormat)

	//Recreated the Current Date/Time object, to removed any UTC/Timezone.
	curDateTime, _ := time.Parse(strStandardFormat, curDateTimeStr)

	//Unix Value for CurrentDateTime.
	NowTimeUnix := curDateTime.Unix()

	//Create the Inputted Date/Time object, the created object will without UTC/Timezone
	InputTime, _ := time.Parse(strStandardFormat, ValDate+"T"+ValTime+":00")

	//Unix Value for Inputted DateTime.
	InputTimeUnix := InputTime.Unix()

	//Compare and return the result
	return NowTimeUnix > InputTimeUnix
}

//ValidateInputDateMaxYesterday - a function to validate the input date not greater than yesterday.
func ValidateInputDateMaxYesterday(ValDate string) (validatedResult bool) {
	//Get Current Date/Time
	currentTime := time.Now()

	//Get Formatted Current Date Time
	curDateStr := currentTime.Format(strStandardFormat)

	//Recreated the Current Date/Time object, to removed any UTC/Timezone.
	curDateTime, _ := time.Parse(strStandardFormat, curDateStr)

	//Unix Value for CurrentDateTime.
	NowDateUnix := curDateTime.Unix()

	//Create the Inputted Date/Time object, the created object will without UTC/Timezone
	InputDate, _ := time.Parse(strStandardFormat, ValDate)

	//Unix Value for Inputted DateTime.
	InputDateUnix := InputDate.Unix()

	//Compare and return the result
	//fmt.Printf("%v | %v : %v | %v", curDateStr, ValDate, NowDateUnix, InputDateUnix)
	return NowDateUnix > InputDateUnix

}
