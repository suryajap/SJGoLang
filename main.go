package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/suryajap/SJGoLang/SGAirTemp"
)

func main() {
	//Create Database Connection Placeholder.
	DBConn, err := SGAirTemp.InitDBConn("sqlite3", "sg-airtemp.db")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	DBConn.PrepareDBTable()

	fmt.Println("Please choose:")
	fmt.Println("1. Print Recorded Stations")
	fmt.Println("2. Print Recorded Temperature Readings Order by Stations")
	fmt.Println("3. Get the Temperature Recording based on User Inputted Date/Time")
	fmt.Println("4. Get the 1 Day Temperature Statistic (Hourly Recording for 24 Hour)")
	fmt.Println("5. Get the 1 Month Temperature Statistic (Hourly Recording for 24 Hours each day - max is Last Month)")
	fmt.Println("6. Get the statistic from all the saved data")
	fmt.Println("7. Get the statistic for 1 FULL day of data")

	//Get the Input of Date from user.
	inpChoiceValStr := SGAirTemp.GetUserInput("\nYour Choice: ")
	inpChoiceValInt, _ := strconv.Atoi(inpChoiceValStr)

	switch inpChoiceValInt {
	case 1:
		DBConn.PrintStations()
	case 2:
		DBConn.PrintTemperatureReading("", "")
	case 3:
		errMsg := DBConn.UserInputAndSaveTemperatureData()
		if errMsg != "" {
			log.Fatal(errMsg)
			os.Exit(1)
		}
	case 4:
		errMsg := DBConn.GetOneDayStatistic("")
		if errMsg != "" {
			log.Fatal(errMsg)
			os.Exit(1)
		}
	case 5:
		errMsg := DBConn.GetOneMonthStatistic()
		if errMsg != "" {
			log.Fatal(errMsg)
			os.Exit(1)
		}
	case 6:
		errMsg := DBConn.GetAllDataStatistic()
		if errMsg != "" {
			log.Fatal(errMsg)
			os.Exit(1)
		}
	case 7:
		errMsg := DBConn.GetOneFullDayStatistic()
		if errMsg != "" {
			log.Fatal(errMsg)
			os.Exit(1)
		}
	default:
		fmt.Println("Please choose valid option.")
	}
}
