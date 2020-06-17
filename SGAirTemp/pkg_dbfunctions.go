package SGAirTemp

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
)

//InitDBConn function to initiate the connection and store it to the global DB connection.
func InitDBConn(host, port string) (*DB, error) {
	db, err := sql.Open(host, port)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

//PrepareDBTable function to prepare the database if not exists
func (dbc *DB) PrepareDBTable() {
	//Make sure the table is exists, especially for the new environment.
	//Create table for stations
	statement, _ := dbc.Prepare("CREATE TABLE IF NOT EXISTS stations (station_id TEXT PRIMARY KEY, station_name TEXT, loc_latitude TEXT, loc_longitude TEXT)")
	statement.Exec()
	//Create table for readings
	statement, _ = dbc.Prepare("CREATE TABLE IF NOT EXISTS readings (station_id TEXT, yr TEXT, mo TEXT, dt TEXT, hr TEXT, mi TEXT, value REAL)")
	statement.Exec()

}

//InsertTemperatureReading - a function to check if the wheather reading grabbed from API exists on the Database and save it if we can't find it.
func (dbc *DB) InsertTemperatureReading(stationID, timeStamp string, Value float64) {

	//Parse the timestamp. Ugly but more straight-forward
	//Start - TODO: research better method?
	yr := string([]rune(timeStamp)[0:4])
	mo := string([]rune(timeStamp)[5:7])
	dt := string([]rune(timeStamp)[8:10])
	hr := string([]rune(timeStamp)[11:13])
	mi := string([]rune(timeStamp)[14:16])
	//End - TODO: research better method?
	totalRow, _ := strconv.Atoi(dbc.GetScalar(fmt.Sprintf("SELECT COUNT(station_id) scalarRes FROM readings WHERE station_id ='%v' AND yr ='%v' AND mo ='%v' AND dt ='%v' AND hr ='%v' AND mi ='%v'", stationID, yr, mo, dt, hr, mi)))

	//No Data found for this Temperature ID, add it.
	if totalRow <= 0 {
		statement, err := dbc.Prepare("INSERT INTO readings(station_id, yr, mo, dt, hr, mi, value) VALUES(?, ?, ?, ?, ?, ?, ?)")
		statement.Exec(stationID, yr, mo, dt, hr, mi, strconv.FormatFloat(Value, 'f', 5, 64))
		if err != nil {
			fmt.Printf("\nError During Insert:%v", err)
		}
	}
}

//InsertStation - a function to check if the station grabbed from API is exist on the Datbase and save it if we can't find it.
func (dbc *DB) InsertStation(stationID, StationName, LocLatitude, LocLongitude string) {
	totalRow, _ := strconv.Atoi(dbc.GetScalar(fmt.Sprintf("SELECT COUNT(station_id) scalarRes FROM stations WHERE station_id ='%v'", stationID)))

	//No Data found for this station ID, add it.
	if totalRow <= 0 {
		statement, err := dbc.Prepare("INSERT INTO stations(station_id, station_name, loc_latitude, loc_longitude) VALUES(?, ?, ?, ?)")
		statement.Exec(stationID, StationName, LocLatitude, LocLongitude)
		if err != nil {
			fmt.Printf("\nError During Insert:%v", err)
		}
	}
}

//GetScalar - a function to return the scalar value of a query with the scalarRes as the return name.
//Ideal for SELECT COUNT(1) scalarRes
func (dbc *DB) GetScalar(strSQL string) (scalarRes string) {
	rows, err := dbc.Query(strSQL)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		rows.Scan(&scalarRes)
	}
	return scalarRes
}

//PrintStations - function to print the Stations to the console
func (dbc *DB) PrintStations() {
	MaxStationNameLen := dbc.GetScalar("SELECT LENGTH(station_name) scalarRes FROM stations ORDER BY LENGTH(station_name) DESC LIMIT 1")

	totalRow, _ := strconv.Atoi(dbc.GetScalar("SELECT COUNT(station_id) scalarRes FROM stations"))

	if totalRow > 0 {
		rows, _ := dbc.Query("SELECT station_id, station_name, loc_latitude, loc_longitude FROM stations")
		var StationID string
		var stationName string
		var locLatitude string
		var locLongitude string
		fmt.Printf("%9s"+" | "+"%"+MaxStationNameLen+"s"+" | %9s | %9s\n", "StationID", "StationName", "Latitude", "Longitude")
		MaxStationNameLenInt, _ := strconv.Atoi(MaxStationNameLen)
		fmt.Printf("%s\n", strings.Repeat("=", MaxStationNameLenInt+36))

		for rows.Next() {
			rows.Scan(&StationID, &stationName, &locLatitude, &locLongitude)
			fmt.Printf("%9s"+" | "+"%"+MaxStationNameLen+"s"+" | %9s | %9s\n", StationID, stationName, locLatitude, locLongitude)
		}
	} else {
		fmt.Printf("No Station being found, please retrive it fromt the API")
	}

}

//PrintTemperatureReading - function to print the Temperature Reading to the console
func (dbc *DB) PrintTemperatureReading(dateVal, timeVal string) {
	whereCondSQL := ""
	//Flexible array, by using splices
	WhereCondition := []string{}
	//Make sure only
	if len(dateVal) > 0 {
		dateVal = CheckInputDate(dateVal)
		arrDate := strings.Split(dateVal, "-")
		yrCond := arrDate[0]
		moCond := arrDate[1]
		dtCond := arrDate[2]
		WhereCondition = append(WhereCondition, fmt.Sprintf("yr ='%v'", yrCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("mo ='%v'", moCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("dt ='%v'", dtCond))
	}

	if len(timeVal) == 5 {
		timeVal = CheckInputTime(timeVal)
		arrTime := strings.Split(timeVal, ":")
		hrCond := arrTime[0]
		miCond := arrTime[1]
		WhereCondition = append(WhereCondition, fmt.Sprintf("hr ='%v'", hrCond))
		WhereCondition = append(WhereCondition, fmt.Sprintf("mi ='%v'", miCond))
	}

	if len(WhereCondition) > 0 {
		whereCondSQL = fmt.Sprintf(" WHERE %v ", strings.Join(WhereCondition[:], " AND "))
	}

	MaxStationNameLen := dbc.GetScalar("SELECT LENGTH(station_name) scalarRes FROM stations ORDER BY LENGTH(station_name) DESC LIMIT 1")
	totalRow, _ := strconv.Atoi(dbc.GetScalar(fmt.Sprintf("SELECT COUNT(station_id) scalarRes FROM readings %v", whereCondSQL)))

	if totalRow > 0 {
		rows, _ := dbc.Query(fmt.Sprintf("SELECT s.station_name, yr, mo, dt, hr, mi, value FROM readings r INNER JOIN stations s ON s.station_id = r.station_id %v ORDER BY s.station_name, yr, mo, dt, hr, mi", whereCondSQL))
		var StationName string
		var yr string
		var mo string
		var dt string
		var hr string
		var mi string
		var value string

		fmt.Printf("\n%"+MaxStationNameLen+"s"+" | %16s | %s\n", "StationName", "Date/Time", "Value")
		MaxStationNameLenInt, _ := strconv.Atoi(MaxStationNameLen)
		fmt.Printf("%s\n", strings.Repeat("=", MaxStationNameLenInt+27))
		for rows.Next() {
			rows.Scan(&StationName, &yr, &mo, &dt, &hr, &mi, &value)
			fmt.Printf("%"+MaxStationNameLen+"s"+" | %v-%v-%v %v:%v | %v\n", StationName, yr, mo, dt, hr, mi, value)
		}
	} else {
		fmt.Printf("No Reading being found, please retrive it from the API")
	}

}
