package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"runtime"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

var USER, PASSWORD string

var (
	deptNo    string
	deptName  string
	empNo     int
	fromDate  string
	toDate    string
	birthDate string
	firstName string
	lastName  string
	gender    string
	hireDate  string
	salary    int
	title     string
)

var operationPtr = flag.String("operator", "", "Operation: SELECT, DELETE, UPDATE, INSERT")
var countPtr = flag.Int("count", 1, "Repeat: Number of times to repeat the benchmark.")
var dbPtr = flag.String("db", "", "Database: Name of the DB to perform operations on.")
var tablePtr = flag.String("table", "", "Table: Name of the table to perform operations on.")
var conditionPtr = flag.String("condition", "", "Condition: Constraint on the transaction.")
var jsonConfig = flag.String("config", "", "Configuration: Input a predefined configuration file.")

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Benchmark_initializeDB(bench *testing.B) {
	for iter := 0; iter < bench.N; iter++ {
		db := initializeDB()
		db.Close()
	}
}

func Benchmark_prepareStatement(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	for iter := 0; iter < bench.N; iter++ {
		rows := prepareStatement(db, *operationPtr, *tablePtr, *conditionPtr)
		rows.Close()
	}
}

func Benchmark_processData(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	rows := prepareStatement(db, *operationPtr, *tablePtr, *conditionPtr)
	defer rows.Close()

	for iter := 0; iter < bench.N; iter++ {
		_ = processData(rows)
	}
}

func configParse(inputFile ...string) {
	var config map[string]interface{}

	if inputFile != nil {
		file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", inputFile[0]))
		if err != nil {
			log.Fatal(fmt.Sprintf("file io error: %s\n", err))
		}
		json.Unmarshal(file, &config)
	} else {
		file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", *jsonConfig))
		if err != nil {
			log.Fatal(fmt.Sprintf("file io error: %s\n", err))
		}
		json.Unmarshal(file, &config)
	}

	USER = config["user"].(string)
	PASSWORD = config["pass"].(string)
}

func validateInput() {
	if *tablePtr == "" {
		log.Fatal("Please specify a MySQL table using the --table option.")
	} else if *dbPtr == "" {
		log.Fatal("Please specify a MySQL database using the --database option.")
	} else if *operationPtr == "" {
		log.Fatal("Please specify a MySQL operation using the --operator option.")
	} else if *jsonConfig == "" {
		log.Fatal("Please specify a configuration file using the --config option.")
	}
}

func initializeDB(inputParams ...string) *sql.DB {
	// lomax_test.go uses custom command function name for testing purposes only
	if len(inputParams) != 0 {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", inputParams[0], inputParams[1], inputParams[2]))
		if err != nil {
			log.Fatal(err)
		}
		return db
	} else {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", USER, PASSWORD, *dbPtr))
		if err != nil {
			log.Fatal(err)
		}
		return db
	}
}

func prepareStatement(db *sql.DB, operationPtr string, tablePtr string, conditionPtr string) *sql.Rows {
	stmtOut, err := db.Prepare(fmt.Sprintf("%s FROM %s %s", operationPtr, tablePtr, conditionPtr))
	if err != nil {
		log.Fatal(err)
	}
	rows, err := stmtOut.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer stmtOut.Close()
	return rows
}

func processData(rows *sql.Rows, inputParams ...string) bool {
	if inputParams != nil {
		*tablePtr = inputParams[0]
	}

	for rows.Next() {
		switch *tablePtr {
		case "employees":
			err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate)
			if err != nil {
				log.Fatal(err)
			}
			// log.Println(empNo, birthDate, firstName, lastName, gender, hireDate)
		case "departments":
			err := rows.Scan(&deptNo, &deptName)
			if err != nil {
				log.Fatal(err)
			}
			// log.Println(deptNo, deptName)

		default:
			log.Fatal("Invalid table specified, please check the --table option.")
			return false
		}
		err := rows.Err()
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	// Only reaches here if rows is empty.
	if rows != nil {
		return true
	}
	return false
}

func runBenchmarks() {
	br := testing.Benchmark(Benchmark_initializeDB)
	fmt.Println(fmt.Sprintf("[%s    ]: Time Taken: %s 	Ops: %s", GetFunctionName(Benchmark_initializeDB), br.T, br))

	br = testing.Benchmark(Benchmark_prepareStatement)
	fmt.Println(fmt.Sprintf("[%s]: Time Taken: %s 	Ops: %s", GetFunctionName(Benchmark_prepareStatement), br.T, br))

	br = testing.Benchmark(Benchmark_processData)
	fmt.Println(fmt.Sprintf("[%s     ]: Time Taken: %s 	Ops: %s", GetFunctionName(Benchmark_processData), br.T, br))
}

func main() {
	flag.Parse()

	validateInput()

	configParse()

	runBenchmarks()
}
