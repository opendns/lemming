// The lemming/lomax is a benchmarking tool for the lemming suite of
// MySQL Data appliations at OpenDNS. Lomax aims to support various
// forks and flavors of MySQL for benchmark and testing.
//
// Example:
//
//	   ./lomax --vector=openstack-generic-test-select.json --config=openstack-generic-config.json
//
//		[main.BenchmarkInitializeDB    ]: Time Taken: 2.995021859s      Ops:   200000        14975 ns/op
//		[main.BenchmarkPrepareStatement]: Time Taken: 1.394912791s      Ops:     5000       278982 ns/op
//		[main.BenchmarkProcessData     ]: Time Taken: 1.214483332s      Ops: 100000000          12.1 ns/op

package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/opendns/lemming/lib/log"
)

// USER : The MySQL user, passed in through the config file
var USER string

// PASSWORD : The MySQL user's password, passed in through the config file
var PASSWORD string

// This schema is only valid for datacharmer/test_db
// If you would like to use your own, please change accordingly.
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

// Make stuff that is common globally accessible
var operationPtr, columnsPtr, dbPtr, tablePtr, conditionPtr interface{}
var testVectorConfig = flag.String("vector", "", "Test Vectors: Input a predefined test vector configuration file.")
var jsonConfig = flag.String("config", "", "Configuration: Input a predefined configuration file.")
var graphDumps = flag.String("graph", "", "Output Dumps: If defined, the program outputs the data to a specified file format.")
var countPtr = flag.Int("count", 1, "Repeat: Number of times to repeat the benchmark.")
var logPrefix = flag.String("logprefix", "", "Log: If defined the logs are prefixed with this name")
var config map[string]interface{}
var benchmarkData []string

func initPtrs() {
	if testVectorConfig != nil {
		operationPtr = config["action"]
		columnsPtr = config["columns"]
		dbPtr = config["test_db"]
		tablePtr = config["test_table"]
		conditionPtr = config["condition"]
	} else {
		operationPtr = flag.String("operator", "", "Operation: SELECT, DELETE, UPDATE, INSERT")
		dbPtr = flag.String("db", "", "Database: Name of the DB to perform operations on.")
		tablePtr = flag.String("table", "", "Table: Name of the table to perform operations on.")
		conditionPtr = flag.String("condition", "", "Condition: Constraint on the transaction.")
	}
}

// GetFunctionName : Returns the name of the passed function
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// BenchmarkInitializeDB : Benchmark helper function to benchmark initializeDB()
func BenchmarkInitializeDB(bench *testing.B) {
	for iter := 0; iter < bench.N; iter++ {
		db := initializeDB()
		db.Close()
	}
}

// BenchmarkPrepareStatement : Benchmark helper function to benchmark prepareStatement()
func BenchmarkPrepareStatement(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	for iter := 0; iter < bench.N; iter++ {
		if testVectorConfig != nil {
			rows := prepareStatement(db, config["action"].(string), config["columns"].(string), config["test_table"].(string), config["condition"].(string))
			rows.Close()
		} else {
			rows := prepareStatement(db, operationPtr.(string), columnsPtr.(string), tablePtr.(string), conditionPtr.(string))
			rows.Close()
		}
	}
}

// BenchmarkProcessData : Benchmark helper function to benchmark processData()
func BenchmarkProcessData(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	rows := prepareStatement(db, operationPtr.(string), columnsPtr.(string), tablePtr.(string), conditionPtr.(string))
	defer rows.Close()

	for iter := 0; iter < bench.N; iter++ {
		_ = processData(rows)
	}
}

func configParse(inputFile ...string) {

	if inputFile != nil {
		file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", inputFile[0]))
		if err != nil {
			log.Error(fmt.Sprintf("File IO Error: %s\n", err.Error()))
		}
		fileTestConfig, errTestConfig := ioutil.ReadFile(fmt.Sprintf("./testvectors/%s", inputFile[1]))
		if errTestConfig != nil {
			log.Error(fmt.Sprintf("Test config File IO Error: %s\n", err.Error()))
		}
		json.Unmarshal(file, &config)
		json.Unmarshal(fileTestConfig, &config)
	} else {
		file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", *jsonConfig))
		if err != nil {
			log.Error(fmt.Sprintf("File IO Error: %s\n", err.Error()))
		}
		fileTestConfig, errTestConfig := ioutil.ReadFile(fmt.Sprintf("./testvectors/%s", *testVectorConfig))
		if errTestConfig != nil {
			log.Error(fmt.Sprintf("Test config File IO Error: %s\n", err.Error()))
		}
		json.Unmarshal(file, &config)
		json.Unmarshal(fileTestConfig, &config)
	}

	USER = config["user"].(string)
	PASSWORD = config["pass"].(string)
}

func validateInput() {
	if tablePtr == "" {
		log.Error("Please specify a MySQL table using the --table option.")
	} else if dbPtr == "" {
		log.Error("Please specify a MySQL database using the --database option.")
	} else if operationPtr == "" && *testVectorConfig == "" {
		log.Error("Please specify a MySQL operation using the --operator option or specify a test vector using --vector option.")
	} else if *jsonConfig == "" {
		log.Error("Please specify a configuration file using the --config option.")
	}
}

func initializeDB(inputParams ...string) *sql.DB {
	// lomax_test.go uses custom command function name for testing purposes only
	if len(inputParams) != 0 {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", inputParams[0], inputParams[1], inputParams[2]))
		if err != nil {
			log.Error(err.Error())
		}
		return db
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", USER, PASSWORD, dbPtr.(string)))
	if err != nil {
		log.Error(err.Error())
	}
	return db
}

func prepareStatement(db *sql.DB, operationPtr string, columnsPtr string, tablePtr string, conditionPtr string) *sql.Rows {
	stmtOut, err := db.Prepare(fmt.Sprintf("%s %s FROM %s %s", operationPtr, columnsPtr, tablePtr, conditionPtr))
	if err != nil {
		log.Error(err.Error())
	}
	rows, err := stmtOut.Query()
	if err != nil {
		log.Error(err.Error())
	}
	defer stmtOut.Close()
	return rows
}

func processData(rows *sql.Rows, inputParams ...string) bool {
	if inputParams != nil {
		tablePtr = inputParams[0]
	}

	for rows.Next() {
		switch tablePtr {
		case "employees":
			err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate)
			if err != nil {
				log.Error(err.Error())
			}
			// log.Debug(strconv.Itoa(empNo), birthDate, firstName, lastName, gender, hireDate)
		case "departments":
			err := rows.Scan(&deptNo, &deptName)
			if err != nil {
				log.Error(err.Error())
			}
			// log.Debug(strconv.Itoa(deptNo), deptName)

		default:
			log.Error("Invalid table specified, please check the --table option.")
			return false
		}
		err := rows.Err()
		if err != nil {
			log.Error(err.Error())
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
	if *logPrefix == "" {
		log.Warning(fmt.Sprintf("[%s]: No --logprefix defined, log file will NOT be created", GetFunctionName(exportData)))
	}

	br := testing.Benchmark(BenchmarkInitializeDB)
	collectData(br, BenchmarkInitializeDB)

	br = testing.Benchmark(BenchmarkPrepareStatement)
	collectData(br, BenchmarkPrepareStatement)

	br = testing.Benchmark(BenchmarkProcessData)
	collectData(br, BenchmarkProcessData)
}

func writeToFile() *os.File {
	filePtr, err := os.OpenFile(fmt.Sprintf("./results/%s.%s", *logPrefix, *graphDumps), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Error("[%s]: Cannot create file for writing.", GetFunctionName(writeToFile))
	}
	return filePtr
}

func collectData(br testing.BenchmarkResult, funcPtr func(*testing.B)) {
	if *graphDumps == "json" && *logPrefix != "" {
		benchmarkStr := fmt.Sprintf("[%s]: Time Taken: %s, Ops: %s", GetFunctionName(funcPtr), br.T, br)
		fmt.Println(benchmarkStr)
		benchmarkData = append(benchmarkData, string(benchmarkStr))
	} else if *graphDumps == "csv" && *logPrefix != "" {
		benchmarkStr := fmt.Sprintf("[%s]: Time Taken: %s, Ops: %s", GetFunctionName(funcPtr), br.T, br)
		fmt.Println(benchmarkStr)
		benchmarkData = append(benchmarkData, string(benchmarkStr))
	} else {
		fmt.Println(fmt.Sprintf("[%s]: Time Taken: %s, Ops: %s", GetFunctionName(funcPtr), br.T, br))
	}
}

func exportData() {
	if *graphDumps == "json" && *logPrefix != "" {
		filePtr := writeToFile()
		jsonString, _ := json.MarshalIndent(benchmarkData, "", "  ")
		_, err := filePtr.WriteString(string(jsonString))
		if err != nil {
			log.Error("[%s]: Couldn't write to the JSON output file")
		}
		defer filePtr.Close()
	} else if *graphDumps == "csv" && *logPrefix != "" {
		filePtr := writeToFile()
		csvWriter := csv.NewWriter(filePtr)
		err := csvWriter.Write(benchmarkData)
		if err != nil {
			log.Error("[%s]: Cannot write to CSV file", err)
		}
		csvWriter.Flush()
		defer filePtr.Close()
	}
}

func main() {
	flag.Parse()

	validateInput()

	configParse()

	initPtrs()

	runBenchmarks()

	if *logPrefix != "" {
		exportData()
	}
}
