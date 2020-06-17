package SGAirTemp

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//DB struct - a placeholder for Database connection.
type DB struct {
	*sql.DB
}

//EarliestDataAvail - Taken form "Coverage" https://data.gov.sg/dataset/realtime-weather-readings
//Must be translated to YYYY-MM-DD
const EarliestDataAvail = "2016-12-14"
const strStandardFormat = "2006-01-02"

//Location struct - to form the location object inside the Station
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

//Station struct - to form the station object
type Station struct {
	StationID   string   `json:"device_id"`
	StationName string   `json:"name"`
	Location    Location `json:"location"`
}

//Metadata struct - to form the location object inside the Station
type Metadata struct {
	Station []Station `json:"stations"`
}

//TemperatureReading - to as the placeholder to stored the TemperatureReading
type TemperatureReading struct {
	StationID string  `json:"station_id"`
	Value     float64 `json:"value"`
}

//TemperatureData struct - as placeholder for the response: items.readings
type TemperatureData struct {
	Timestamp          string               `json:"timestamp"`
	TemperatureReading []TemperatureReading `json:"readings"`
}

//TemperatureResponse struct - response body for the TemperatureAPI response
type TemperatureResponse struct {
	Metadata        Metadata          `json:"metadata"`
	TemperatureData []TemperatureData `json:"items"`
}

//GetDateInput - Get Date input from user.
func GetDateInput() string {
	//Get the Input of Date from user.
	inpDate := bufio.NewReader(os.Stdin)
	fmt.Print("Provide Date (format: YYYY-MM-DD): ")
	dateVal, _ := inpDate.ReadString('\n')

	//Validate the Date Input
	dateVal = CheckInputDate(dateVal)

	//All validation is completed and pass, return it.
	return dateVal
}

//GetTimeInput - Get Time input from user.
func GetTimeInput() string {
	//Get the time input from the user.
	inpTime := bufio.NewReader(os.Stdin)
	fmt.Print("Provide Time (format: HH:mm): ")
	timeVal, _ := inpTime.ReadString('\n')

	//Validate the Time Input
	timeVal = CheckInputTime(timeVal)

	//All validation is completed and pass, return it.
	return timeVal
}

//GetDateTimeInput - Get Date/Time input from user.
//Validate if the Date/Time is valid.
//If Date/Time is empty, it wil use the current Date/Time
func GetDateTimeInput() (string, string) {

	//Get the date input from the user.
	dateVal := GetDateInput()

	//Get the time input from the user.
	timeVal := GetTimeInput()

	return dateVal, timeVal
}

//APICallAndGetResponse - API Call for the
func APICallAndGetResponse(ValDate, ValTime string) (response TemperatureResponse, errorMessage string) {
	//String API Call, we only need to get the data until the minute level.
	strAPICall := "https://api.data.gov.sg/v1/environment/air-temperature?date_time=" + url.QueryEscape(ValDate+"T"+ValTime+":00") + "&date=" + ValDate
	res, err := http.Get(strAPICall)

	if err != nil {
		log.Fatal(err)
	} else {
		//We found the
		if res.StatusCode == 200 {
			//Read the response body and stored it to variable body
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err.Error())
			} else {
				//API couldn't provide data.
				//Sometimes the API don't have the data for that day, at least I found the date they can't provide data is: 2020-06-10 and 2020-06-11
				if len(body) <= 110 {
					errorMessage = fmt.Sprintf("Error, can't find API response body info for %v %v\nDuring calling-> %v", ValDate, ValTime, strAPICall)
				} else {
					//All fine, Unmarshal the response from the API Response Body - Parsing
					json.Unmarshal([]byte(body), &response)
				}
			}
		} else {
			errorMessage = fmt.Sprintf("\nError Code: %v\nDuring calling-> %v", res.StatusCode, strAPICall)
		}
	}
	return response, errorMessage
}

//CallTemperatureAPIAndSave - Get User Input for Date and Time and Save it to Database
func (dbc *DB) CallTemperatureAPIAndSave(ValDate, ValTime string, displayResult bool) (resultInfo string) {
	Stations := []string{}
	//Call the API for TemperatureReading and retrieve the response.
	response, err1 := APICallAndGetResponse(ValDate, ValTime)

	if err1 != "" {
		resultInfo = err1
	} else {
		//Iterate for the Station object and save it to the Database
		for i := 0; i < (len(response.Metadata.Station)); i++ {
			st := response.Metadata.Station[i]
			Stations = append(Stations, st.StationID, st.StationName)
			dbc.InsertStation(st.StationID, st.StationName, strconv.FormatFloat(st.Location.Latitude, 'f', 5, 64), strconv.FormatFloat(st.Location.Longitude, 'f', 5, 64))
		}

		//Iterate for the TemperatureReading object and save it to the Database
		TemperatureData := &response.TemperatureData[0]
		strTimeStamp := TemperatureData.Timestamp
		for i := 0; i < (len(TemperatureData.TemperatureReading)); i++ {
			rd := TemperatureData.TemperatureReading[i]
			//fmt.Printf("%"+MaxStationNameLen+"s"+" | %v | %v\n", Stations[rd.StationID], strTimeStamp, rd.Value)
			dbc.InsertTemperatureReading(rd.StationID, strTimeStamp, rd.Value)
		}
		if displayResult == true && len(TemperatureData.TemperatureReading) > 0 {
			fmt.Printf("\nThe Result returned by the API might not be the same timing as what you input.\nThe API will sometimes return the nearest time on what you requested.")
			arrDateTime := strings.Split(strTimeStamp, "T")
			valDate := string(arrDateTime[0])
			valTime := string([]rune(arrDateTime[1])[0:5])
			dbc.PrintTemperatureReading(valDate, valTime)
		}

		fmt.Printf("%v", strTimeStamp)

	}
	return resultInfo
}

//UserInputAndSaveTemperatureData - Get User Input for Date and Time and Save it to Database
func (dbc *DB) UserInputAndSaveTemperatureData() string {
	resultInfo := ""
	dateVal, timeVal := GetDateTimeInput()

	if ValidateDateTimeLessThanNow(dateVal, timeVal) == true {
		resultInfo = dbc.CallTemperatureAPIAndSave(dateVal, timeVal, true)
	} else {
		resultInfo = fmt.Sprintf("The Inputted date/time '%v %v' must be not later than current time ", dateVal, timeVal)
	}
	return resultInfo
}

//GetChoosenStation - function to get ID of chosen Station, empty string if the user choose all.
func (dbc *DB) GetChoosenStation() string {
	StringChosenID := ""
	StationIDs := []string{}
	var StationID, StationName string
	var i int
	StrQuery := "SELECT station_id, station_name FROM stations ORDER BY station_name"
	rows, _ := dbc.Query(StrQuery)

	fmt.Println("Please choose the Station:")
	fmt.Println("0. ALL Stations")
	for rows.Next() {
		rows.Scan(&StationID, &StationName)
		//Add StationID to the splice.
		StationIDs = append(StationIDs, StationID)
		i++
		fmt.Printf("\n%v. %v", i, StationName)
	}

	if i > 0 {
		//Get the Input from the User
		inp := GetUserInput(fmt.Sprintf("\nYour choice (0-%v): ", i))
		inputInt, _ := strconv.Atoi(inp)
		if inputInt > 0 {
			//Minus 1 since the array start from 1.
			StringChosenID = StationIDs[inputInt-1]
		} else {
			fmt.Printf("\nThe input date is not between 1 to %v, so ALL Station is selected by default.", i)
		}
	}
	return StringChosenID
}

//GetOneDayStatistic a function to get the hourly data
func (dbc *DB) GetOneDayStatistic(strDateRequested string) string {
	resultInfo := ""
	var StationName string
	var yr string
	var mo string
	var dt string
	var hr string
	var mi string
	var value float64
	var cntExisting int

	dontShowMessage := false
	var dateVal string
	if len(strDateRequested) > 0 {
		dateVal = CheckInputDate(strDateRequested)
		dontShowMessage = true
	}

	callAPI := false

	minTemp := 9999.99
	maxTemp := -9999.99
	totalTemp := 0.00
	totalReadings := 0.00
	centerData := 0

	//Minimum Temperature Date/Time and StationName
	minTempDateTimeAndStation := ""
	//Maximum Temperature Date/Time and StationName
	maxTempDateTimeAndStation := ""

	//Flexible array, by using splices
	WhereCondition := []string{}

	//Fixed array. To generate hourly
	ArrayHour := []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}

	if dontShowMessage == false {
		dateVal = GetDateInput()
	}

	if ValidateInputDateMaxYesterday(dateVal) == true {
		StrQuery := ""

		//Get the maximum length of the all the Station's name.
		MaxStationNameLen := dbc.GetScalar("SELECT LENGTH(station_name) scalarRes FROM stations ORDER BY LENGTH(station_name) DESC LIMIT 1")
		MaxStationNameLenInt, _ := strconv.Atoi(MaxStationNameLen)

		//extract the year, month and date
		yrCond := string([]rune(dateVal)[0:4])
		moCond := string([]rune(dateVal)[5:7])
		dtCond := string([]rune(dateVal)[8:10])

		StationID := ""
		if dontShowMessage == false {
			StationID = dbc.GetChoosenStation()
		}

		//Set the Where condition
		WhereCondition = append(WhereCondition, fmt.Sprintf("yr ='%v'", yrCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("mo ='%v'", moCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("dt ='%v'", dtCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("hr IN ('%v')", strings.Join(ArrayHour[:], "','")))
		WhereCondition = append(WhereCondition, "mi ='00'")
		if len(StationID) > 0 {
			WhereCondition = append(WhereCondition, fmt.Sprintf("r.station_id ='%v'", StationID))
		}

		StrQueryCnt := fmt.Sprintf("SELECT count(value) AS scalarRes FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE %v GROUP BY s.station_name", strings.Join(WhereCondition[:], " AND "))
		GetTotalRows, _ := strconv.Atoi(dbc.GetScalar(StrQueryCnt))

		if GetTotalRows > 0 {
			StrQuery := fmt.Sprintf("SELECT s.station_name, count(value) AS cnt FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE %v GROUP BY s.station_name ORDER BY s.station_name", strings.Join(WhereCondition[:], " AND "))
			rows, _ := dbc.Query(StrQuery)
			for rows.Next() {
				rows.Scan(&StationName, &cntExisting)
				if cntExisting < 24 {
					//One of the station(s) not having the 24 data, we need to pull it from the API
					callAPI = true
				}
			}
		} else {
			//Couldn't find any data, call the API
			callAPI = true
		}

		if callAPI == true {
			fmt.Printf("\nCalling the API to check for every hour for date (YYYY-MM-DD): %v\n", dateVal)
			for h := 0; h < 24; h++ {
				if dontShowMessage == false {
					fmt.Printf("%v ", ArrayHour[h]+":00")
				} else {
					fmt.Printf(".")
				}
				resultInfo := dbc.CallTemperatureAPIAndSave(dateVal, ArrayHour[h]+":00", false)
				if resultInfo != "" {
					fmt.Println(resultInfo)
				}
			}
		}

		if dontShowMessage == false {
			StrQueryCnt = fmt.Sprintf("SELECT count(value) AS scalarRes FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE %v ", strings.Join(WhereCondition[:], " AND "))
			GetTotalRows, _ = strconv.Atoi(dbc.GetScalar(StrQueryCnt))

			if GetTotalRows > 0 {
				centerData = int(GetTotalRows / 2)
				EvenPosStart := centerData
				EvenPosEnd := centerData + 1
				EvenPosStartValue := 0.00
				EvenPosEndValue := 0.00

				StrQuery = fmt.Sprintf("SELECT s.station_name, yr, mo, dt, hr, mi, value FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE %v ORDER BY r.value, s.station_name, yr, mo, dt, hr, mi", strings.Join(WhereCondition[:], " AND "))
				rows, _ := dbc.Query(StrQuery)

				fmt.Printf("\n%"+MaxStationNameLen+"s"+" | %16s | %s\n", "StationName", "Date/Time", "Value")
				fmt.Printf("%s\n", strings.Repeat("=", MaxStationNameLenInt+27))

				for rows.Next() {
					rows.Scan(&StationName, &yr, &mo, &dt, &hr, &mi, &value)
					fmt.Printf("%"+MaxStationNameLen+"s"+" | %v-%v-%v %v:%v | %v\n", StationName, yr, mo, dt, hr, mi, value)
					if value > maxTemp {
						maxTemp = value

						//Maximum Temperature Date/Time and StationName
						maxTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
					} else if value == maxTemp {
						maxTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", maxTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
					}

					if value < minTemp {
						minTemp = value

						//Minimum Temperature Date/Time and StationName
						minTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
					} else if value == minTemp {
						maxTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", minTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
					}

					totalTemp += value
					totalReadings++
					if EvenPosStart == int(totalReadings) {
						EvenPosStartValue = value
					}
					if EvenPosEnd == int(totalReadings) {
						EvenPosEndValue = value
					}
				}
				fmt.Printf("\nTotal Readings                   : %v", totalReadings)
				fmt.Printf("\nAverange Readings                : %.2f", (totalTemp / totalReadings))
				//Is Even, so we need to get the average of the two of the center data
				if (GetTotalRows % 2) == 0 {
					fmt.Printf("\nMean Readings                    : %.2f", (EvenPosStartValue+EvenPosEndValue)/2)
				} else {
					//Is ODD, just get the center/middle data
					fmt.Printf("\nMean Readings                    : %.2f", EvenPosEndValue)
				}

				fmt.Printf("\nMinimum Temperature              : %v", minTemp)
				fmt.Printf("\nMinimum Temperature Occurence(s) : ")
				fmt.Printf("\n%v", minTempDateTimeAndStation)
				fmt.Printf("\nMaximum Temperature              : %v", maxTemp)
				fmt.Printf("\nMaximum Temperature Occurence(s) : ")
				fmt.Printf("\n%v", maxTempDateTimeAndStation)
			} else {
				resultInfo = fmt.Sprintf("Couldn't find the data reading for '%v'. ", dateVal)
			}
		}
	} else {
		resultInfo = fmt.Sprintf("The Inputted date %s must be not later than yesterday. ", dateVal)
	}

	return resultInfo
}

//GetUserInput a function to get the user input.
func GetUserInput(strMessage string) string {
	fmt.Printf(strMessage)
	inpChoice := bufio.NewReader(os.Stdin)
	inp, _ := inpChoice.ReadString('\n')
	inp = strings.TrimRight(inp, "\r\n")
	inp = strings.TrimRight(inp, "\n")
	return inp
}

//GetOneMonthStatistic a function to get the one month statistic.
//If the requested month is this month, it will get the data from 01 until yesterday (D-1).
func (dbc *DB) GetOneMonthStatistic() string {
	resultInfo := ""

	var StationName string
	var yr string
	var mo string
	var dt string
	var hr string
	var mi string
	var value float64
	//	var cntExisting int

	//Fixed array. To generate hourly
	ArrayHour := []string{"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23"}

	//Flexible array, by using splices
	WhereCondition := []string{}

	minTemp := 9999.99
	maxTemp := -9999.99
	totalTemp := 0.00
	totalReadings := 0.00
	EvenPosStartValue := 0.00
	EvenPosEndValue := 0.00

	centerData := 0

	//Minimum Temperature Date/Time and StationName
	minTempDateTimeAndStation := ""
	//Maximum Temperature Date/Time and StationName
	maxTempDateTimeAndStation := ""

	strYearMonthInput := GetUserInput(fmt.Sprintf("\nYour input (YYYY-MM): "))
	arrYearMonth := strings.Split(strYearMonthInput, "-")
	intYearInp, _ := strconv.Atoi(arrYearMonth[0])
	intMonthInp, _ := strconv.Atoi(arrYearMonth[1])

	//Year Month is not valid.
	if (regexp.MustCompile(`\d{4}-\d{2}`).MatchString(strYearMonthInput) == false) || (intMonthInp < 1 && intMonthInp > 13) {
		resultInfo = fmt.Sprintf("The inputed Data '%s' is not match with YYYY-MM format.\n", strYearMonthInput)
	} else {

		strYearMonthMinData := strings.Split(EarliestDataAvail, "-")
		intYearMinData, _ := strconv.Atoi(strYearMonthMinData[0])
		intMonthMinData, _ := strconv.Atoi(strYearMonthMinData[1])

		//Year/Month inputted is less than the minimum data available from API
		if (intYearInp < intYearMinData) || (intYearInp == intYearMinData && intMonthInp < intMonthMinData) {
			resultInfo = fmt.Sprintf("The inputted Year/Month is less than the minimum data available from API (%v).", EarliestDataAvail)
		} else {
			//Create the Inputted Date/Time object, the created object will without UTC/Timezone
			InputDate, _ := time.Parse(strStandardFormat, strYearMonthInput+"-01")
			//Last Day of Month - Add 1 month and minus 1 day.
			LastDayOfMonth := InputDate.AddDate(0, 1, -1)
			LastDate := string([]rune(LastDayOfMonth.Format(strStandardFormat))[8:10])
			LastDateInt, _ := strconv.Atoi(LastDate)
			reqDateStr := ""
			//Create the Date object of the EarliestDataAvail
			EarliestDate, _ := time.Parse(strStandardFormat, EarliestDataAvail)

			//Unix Value for Earliest Date.
			EarliestDateUnix := EarliestDate.Unix()

			for d := 1; d <= LastDateInt; d++ {
				if d < 10 {
					reqDateStr = fmt.Sprintf("%v-0%v", strYearMonthInput, d)
				} else {
					reqDateStr = fmt.Sprintf("%v-%v", strYearMonthInput, d)
				}
				reqDate, _ := time.Parse(strStandardFormat, reqDateStr)
				reqDateUnix := reqDate.Unix()
				if reqDateUnix > EarliestDateUnix {
					_ = dbc.GetOneDayStatistic(reqDateStr)
				}
			}

			StationID := dbc.GetChoosenStation()

			//Set the Where condition
			WhereCondition = append(WhereCondition, fmt.Sprintf("yr ='%v'", intYearInp))
			if intMonthInp < 10 {
				WhereCondition = append(WhereCondition, fmt.Sprintf("mo ='0%v'", intMonthInp))
			} else {
				WhereCondition = append(WhereCondition, fmt.Sprintf("mo ='%v'", intMonthInp))
			}

			WhereCondition = append(WhereCondition, fmt.Sprintf("hr IN ('%s')", strings.Join(ArrayHour[:], "','")))
			WhereCondition = append(WhereCondition, "mi ='00'")
			if len(StationID) > 0 {
				WhereCondition = append(WhereCondition, fmt.Sprintf("r.station_id ='%v'", StationID))
			}

			StrQueryCnt := fmt.Sprintf("SELECT count(value) AS scalarRes FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE %v GROUP BY s.station_name", strings.Join(WhereCondition[:], " AND "))
			GetTotalRows, _ := strconv.Atoi(dbc.GetScalar(StrQueryCnt))
			if GetTotalRows > 0 {
				centerData = int(GetTotalRows / 2)
				EvenPosStart := centerData
				EvenPosEnd := centerData + 1

				StrQuery := "SELECT s.station_name, yr, mo, dt, hr, mi, value FROM readings r INNER JOIN stations s ON s.station_id = r.station_id WHERE " + strings.Join(WhereCondition[:], " AND ") + " ORDER BY r.value, s.station_name, yr, mo, dt, hr, mi"
				rows, _ := dbc.Query(StrQuery)
				for rows.Next() {
					rows.Scan(&StationName, &yr, &mo, &dt, &hr, &mi, &value)
					if value > maxTemp {
						maxTemp = value

						//Maximum Temperature Date/Time and StationName
						maxTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
					} else if value == maxTemp {
						maxTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", maxTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
					}

					if value < minTemp {
						minTemp = value

						//Minimum Temperature Date/Time and StationName
						minTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
					} else if value == minTemp {
						minTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", minTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
					}

					totalTemp += value
					totalReadings++
					if EvenPosStart == int(totalReadings) {
						EvenPosStartValue = value
					}
					if EvenPosEnd == int(totalReadings) {
						EvenPosEndValue = value
					}
				}
				fmt.Printf("\nTotal Readings                   : %v", totalReadings)
				fmt.Printf("\nAverange Readings                : %.2f", (totalTemp / totalReadings))
				//Is Even, so we need to get the average of the two of the center data
				if (GetTotalRows % 2) == 0 {
					fmt.Printf("\nMean Readings                    : %.2f", (EvenPosStartValue+EvenPosEndValue)/2)
				} else {
					//Is ODD, just get the center/middle data
					fmt.Printf("\nMean Readings                    : %.2f", EvenPosEndValue)
				}

				fmt.Printf("\nMinimum Temperature              : %v", minTemp)
				fmt.Printf("\nMinimum Temperature Occurence(s) : ")
				fmt.Printf("\n%v", minTempDateTimeAndStation)
				fmt.Printf("\nMaximum Temperature              : %v", maxTemp)
				fmt.Printf("\nMaximum Temperature Occurence(s) : ")
				fmt.Printf("\n%v", maxTempDateTimeAndStation)
			}
		}
	}
	return resultInfo
}

//GetAllDataStatistic a function to get the hourly data
func (dbc *DB) GetAllDataStatistic() string {
	resultInfo := ""
	WhereCondition := ""
	var StationName string
	var yr string
	var mo string
	var dt string
	var hr string
	var mi string
	var value float64

	minTemp := 9999.99
	maxTemp := -9999.99
	totalTemp := 0.00
	totalReadings := 0.00
	centerData := 0

	//Minimum Temperature Date/Time and StationName
	minTempDateTimeAndStation := ""
	//Maximum Temperature Date/Time and StationName
	maxTempDateTimeAndStation := ""

	StationID := dbc.GetChoosenStation()

	//Set the Where condition
	if len(StationID) > 0 {
		WhereCondition = fmt.Sprintf(" WHERE r.station_id ='%v' ", StationID)
	}

	StrQueryCnt := fmt.Sprintf("SELECT count(value) AS scalarRes FROM readings r INNER JOIN stations s ON s.station_id = r.station_id %v ", WhereCondition)
	GetTotalRows, _ := strconv.Atoi(dbc.GetScalar(StrQueryCnt))

	if GetTotalRows > 0 {
		centerData = int(GetTotalRows / 2)
		EvenPosStart := centerData
		EvenPosEnd := centerData + 1
		EvenPosStartValue := 0.00
		EvenPosEndValue := 0.00

		StrQuery := fmt.Sprintf("SELECT s.station_name, yr, mo, dt, hr, mi, value FROM readings r INNER JOIN stations s ON s.station_id = r.station_id %v ORDER BY r.value, s.station_name, yr, mo, dt, hr, mi", WhereCondition)
		rows, _ := dbc.Query(StrQuery)

		for rows.Next() {
			rows.Scan(&StationName, &yr, &mo, &dt, &hr, &mi, &value)
			if value > maxTemp {
				maxTemp = value

				//Maximum Temperature Date/Time and StationName
				maxTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
			} else if value == maxTemp {
				maxTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", maxTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
			}

			if value < minTemp {
				minTemp = value

				//Minimum Temperature Date/Time and StationName
				minTempDateTimeAndStation = fmt.Sprintf("  - %v-%v-%v %v:%v -> %v\n", yr, mo, dt, hr, mi, StationName)
			} else if value == minTemp {
				minTempDateTimeAndStation = fmt.Sprintf("%v  - %v-%v-%v %v:%v -> %v\n", minTempDateTimeAndStation, yr, mo, dt, hr, mi, StationName)
			}

			totalTemp += value
			totalReadings++
			if EvenPosStart == int(totalReadings) {
				EvenPosStartValue = value
			}
			if EvenPosEnd == int(totalReadings) {
				EvenPosEndValue = value
			}
		}
		fmt.Printf("\nTotal Readings                   : %v", totalReadings)
		fmt.Printf("\nAverange Readings                : %.2f", (totalTemp / totalReadings))
		//Is Even, so we need to get the average of the two of the center data
		if (GetTotalRows % 2) == 0 {
			fmt.Printf("\nMean Readings                    : %.2f", (EvenPosStartValue+EvenPosEndValue)/2)
		} else {
			//Is ODD, just get the center/middle data
			fmt.Printf("\nMean Readings                    : %.2f", EvenPosEndValue)
		}

		fmt.Printf("\nMinimum Temperature              : %v", minTemp)
		fmt.Printf("\nMinimum Temperature Occurence(s) : ")
		fmt.Printf("\n%v", minTempDateTimeAndStation)
		fmt.Printf("\nMaximum Temperature              : %v", maxTemp)
		fmt.Printf("\nMaximum Temperature Occurence(s) : ")
		fmt.Printf("\n%v", maxTempDateTimeAndStation)
	}

	return resultInfo
}
