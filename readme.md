# SGAirTemp

This is a simple usage of using GoLang to Retrieve the SG Temperature records provided by the Singapore Goverment. 

First data available for the Singapore Air Temperature: Dec 14, 2016

This is my first personal project as a way to learn GoLang.

You will realise there was two statistic for 1 day of data, full per minutes data and hourly data for 24 hours. This is to demonstrate the capabilities for the API and at the same time to demonstrate how long it takes by taking every minutes data vs every hour data. 

The performance might be able to be improved by using different database.


## Pre-requisites

In order to run it properly, you will need to download the 3rd party Sqlite database driver from: github.com/mattn/go-sqlite3

Don't forget to adjust the import based on your own folder structure.

The Sqlite is chosen for portabilities, installation for the RDBMS is not needed. Currently, on the plan to use the MySQL/MariaDB for the bigger data handling capacities and concurent users access. Stay tuned!

The Sqlite database file is not provided, but you can just run it, and it will be created automatically.


## Usage

Simply run the main.go there will be some options you can choose.

Please be aware that the requested data (by time) sometimes are not available from the API. Some sample cases I found there was no data for: June 10, 2020 and June 11, 2020.

Most of the cases, if you requested the data for certain time (2020-06-01 15:03), and the data unfortunately not available on that particular timing, the API will returned the nearest timing, ie: 2020-06-01 15:00.

The daily statistic were based on hourly basis of 24 hours data retrieval from API.


## Contributions

Are more than welcome! For small changes feel free to open a pull request. Or you can email me at: surya.jap@gmail.com

## Credits

- CheckDate function as what I loved from PHP: https://github.com/openset/php2go/blob/master/php/checkdate.go
	I have copied it and saved it to one of my file: pkg_phpcheckdate.go Since I need it for my own.
	You can see more useful tools on php2go, especially if you are coming from PHP developer background like myself.

