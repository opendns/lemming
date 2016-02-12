// The lemming/lomax is a benchmarking tool for the lemming suite of
// MySQL Data appliations at OpenDNS. Lomax aims to support various
// forks and flavors of MySQL for benchmark and testing.
//
// Example:
//
//	   ./lomax --vector=openstack-generic-test-select.json --config=openstack-generic-config.json
//
// 	+--------------------------------+--------------+------------+-----------+-----------+
//	|            FUNCTION            |  TIME TAKEN  | ITERATIONS | MEMALLOCS | MEMBYTES  |
//	+--------------------------------+--------------+------------+-----------+-----------+
//	| main.BenchmarkInitializeDB     | 3.03814889s  |     200000 |   2009863 | 151017520 |
//	| main.BenchmarkPrepareStatement | 3.081639984s |      10000 |    262477 |  10824768 |
//	| main.BenchmarkProcessData      | 1.262388201s |  100000000 |        96 |      8224 |
//	+--------------------------------+--------------+------------+-----------+-----------+

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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/enodata/faker"
	_ "github.com/go-sql-driver/mysql"
	"github.com/olekukonko/tablewriter"
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
var operationPtr, flagPtr, randomPtr, columnsPtr, hostNamePtr, portPtr, dbPtr, tablePtr, conditionPtr string
var logType, logPrefix string
var config map[string]interface{}
var benchmarkData []string
var benchBuffer [][]string
var countPtr float64

var jsonConfig, testVectorConfig string

func init() {
	flag.StringVar(&jsonConfig, "config", "", "JSON config: Input a predefined JSON configuration file.")
	flag.StringVar(&testVectorConfig, "vector", "", "Test Vectors: Input a predefined test vector configuration file.")
	flag.StringVar(&logPrefix, "logprefix", "", "Log Prefix: Defines the prefix for output result file.")
	flag.StringVar(&logType, "logtype", "", "Log Prefix: Defines the output format for storing test results.")
	flag.StringVar(&operationPtr, "operation", "", "Query to run: e.g. SELECT, INSERT..")
	flag.StringVar(&flagPtr, "flag", "", "MySQL Query flag specifier: e.g. LOW_PRIORITY, QUICK, IGNORE..")
	flag.StringVar(&randomPtr, "random", "", "Specify the columns for which you want to generate random data.")
	flag.StringVar(&columnsPtr, "cols", "", "Columns to select in a query.")
	flag.StringVar(&hostNamePtr, "hostname", "", "Hostname of the server to connect to.")
	flag.StringVar(&portPtr, "port", "", "Port number of the host to connec to.")
	flag.StringVar(&dbPtr, "db", "", "DB to perform queries on.")
	flag.StringVar(&tablePtr, "table", "", "Table to use for operations.")
	flag.StringVar(&conditionPtr, "condition", "", "Any conditions to enforce on query.")
	flag.Float64Var(&countPtr, "count", 1, "Number of iterations to perform.")
	flag.StringVar(&USER, "user", "", "MySQL username.")
	flag.StringVar(&PASSWORD, "password", "", "MySQL password.")
}

func setPtrs() {
	configParse()
	if jsonConfig != "" {
		USER = config["user"].(string)
		PASSWORD = config["password"].(string)
		hostNamePtr = config["hostname"].(string)
		portPtr = config["port"].(string)
		dbPtr = config["db"].(string)
	}

	if testVectorConfig != "" {
		USER = config["user"].(string)
		PASSWORD = config["password"].(string)
		operationPtr = config["action"].(string)
		flagPtr = config["flag"].(string)
		columnsPtr = config["columns"].(string)
		dbPtr = config["db"].(string)
		tablePtr = config["table"].(string)
		conditionPtr = config["condition"].(string)
		countPtr = config["count"].(float64)
	}
}

// GetFunctionName : Returns the name of the passed function
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// BenchmarkInitializeDB : Benchmark helper function to benchmark initializeDB()
func BenchmarkInitializeDB(bench *testing.B) {
	for iter := 0; iter < bench.N*int(countPtr); iter++ {
		db := initializeDB()
		db.Close()
	}
}

// BenchmarkPrepareStatement : Benchmark helper function to benchmark prepareStatement()
func BenchmarkPrepareStatement(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	for iter := 0; iter < bench.N*int(countPtr); iter++ {
		rows := prepareStatement(db, operationPtr, flagPtr, randomPtr, columnsPtr, tablePtr, conditionPtr)
		if rows != nil {
			rows.Close()
		}
	}
}

// BenchmarkProcessData : Benchmark helper function to benchmark processData()
func BenchmarkProcessData(bench *testing.B) {
	db := initializeDB()
	defer db.Close()

	rows := prepareStatement(db, operationPtr, flagPtr, randomPtr, columnsPtr, tablePtr, conditionPtr)
	if rows != nil {
		defer rows.Close()
		for iter := 0; iter < bench.N*int(countPtr); iter++ {
			_ = processData(rows, columnsPtr, tablePtr)
		}
	}
}

func configParse(inputFile ...string) {

	if inputFile != nil {
		file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", inputFile[0]))
		if err != nil {
			log.Error(fmt.Sprintf("File IO Error: %s\n", err.Error()))
		}

		if err := json.Unmarshal(file, &config); err != nil {
			log.Error(err.Error())
		}
		USER = config["user"].(string)
		PASSWORD = config["password"].(string)
	} else {
		if jsonConfig != "" {
			file, err := ioutil.ReadFile(fmt.Sprintf("./lib/%s", jsonConfig))
			if err != nil {
				log.Error(fmt.Sprintf("File IO Error: %s\n", err.Error()))
			}
			json.Unmarshal(file, &config)
		}

		if testVectorConfig != "" {
			fileTestConfig, errTestConfig := ioutil.ReadFile(fmt.Sprintf("./testvectors/%s", testVectorConfig))
			if errTestConfig != nil {
				log.Error(fmt.Sprintf("Test config File IO Error: %s\n", errTestConfig.Error()))
			}
			json.Unmarshal(fileTestConfig, &config)
		}
	}
}

func validateInput() {
	if tablePtr == "" && testVectorConfig == "" {
		log.Error("Please specify a MySQL table using the --table option.")
	} else if hostNamePtr == "" && testVectorConfig == "" {
		log.Error("Please specify a hostname using the --hostname option.")
	} else if portPtr == "" && testVectorConfig == "" {
		log.Error("Please specify the port number using the --port option.")
	} else if dbPtr == "" && testVectorConfig == "" {
		log.Error("Please specify a MySQL database using the --database option.")
	} else if operationPtr == "" && testVectorConfig == "" {
		log.Error("Please specify a MySQL operation using the --operation option or specify a test vector using --vector option.")
	} else if columnsPtr == "" && testVectorConfig == "" && randomPtr == "" {
		if operationPtr != "UPDATE" && operationPtr != "DELETE" {
			log.Error("Please specify columns to operate on using the --cols option or specify a test vector using the --vector option.")
		}
	} else if USER == "" && jsonConfig == "" {
		log.Error("Please specify a MySQL user using the --user option.")
	} else if PASSWORD == "" && jsonConfig == "" {
		log.Error("Please specify the MySQL password for the user using the --password option.")
	}
}

func initializeDB(inputParams ...string) *sql.DB {
	// lomax_test.go uses custom command function name for testing purposes only
	if len(inputParams) != 0 {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", inputParams[0], inputParams[1], inputParams[2], inputParams[3], inputParams[4]))
		if err != nil {
			log.Error(err.Error())
		}
		return db
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", USER, PASSWORD, hostNamePtr, portPtr, dbPtr))
	if err != nil {
		log.Error(err.Error())
	}
	return db
}

func prepareStatement(db *sql.DB, operationPtr string, flagPtr string, randomPtr string, columnsPtr string, tablePtr string, conditionPtr string) *sql.Rows {
	if randomPtr == "true" {
		switch tablePtr {
		case "employees":
			columnsPtr = fmt.Sprintf("emp_no, birth_date, first_name, last_name, gender, hire_date")
			conditionPtr = fmt.Sprintf("%s, %s, '%s', '%s', '%s', %s", faker.Number().Number(6), faker.Date().Birthday(10, 40).Format("2006-01-02"), strings.Replace(faker.Name().FirstName(), "'", "", -1), strings.Replace(faker.Name().LastName(), "'", "", -1), "F", faker.Date().Forward(0).Format("2006-01-02"))
		case "dept_emp":
			columnsPtr = fmt.Sprintf("emp_no, dept_no, from_date, to_date")
			conditionPtr = fmt.Sprintf("%s, %s, '%s', '%s'", faker.Number().Number(6), faker.Date().Birthday(10, 40).Format("2006-01-02"), faker.Date().Forward(0).Format("2006-01-02"))
		case "salaries":
			columnsPtr = fmt.Sprintf("emp_no, salary, from_date, to_date")
			conditionPtr = fmt.Sprintf("%s, %s, '%s', '%s'", faker.Number().Number(6), faker.Number().Number(6), faker.Date().Birthday(10, 40).Format("2006-01-02"), faker.Date().Forward(0).Format("2006-01-02"))
		case "titles":
			columnsPtr = fmt.Sprintf("emp_no, title, from_date, to_date")
			conditionPtr = fmt.Sprintf("%s, %s, '%s', '%s'", faker.Number().Number(6), faker.Name().Title(), faker.Date().Birthday(10, 40).Format("2006-01-02"), faker.Date().Forward(0).Format("2006-01-02"))
		case "dept_manager":
			columnsPtr = fmt.Sprintf("emp_no, dept_no, from_date, to_date")
			conditionPtr = fmt.Sprintf("%s, %s, '%s', '%s'", faker.Number().Number(6), faker.Number().Number(4), faker.Date().Birthday(10, 40).Format("2006-01-02"), faker.Date().Forward(0).Format("2006-01-02"))
		case "departments":
			columnsPtr = fmt.Sprintf("dept_no, dept_name")
			conditionPtr = fmt.Sprintf("%s, '%s'", faker.Number().Number(4), strings.Replace(faker.Team().Name(), "'", "", -1))
		default:
			log.Error("[%s]: Invalid SQL table specified. Please check the --table option.", GetFunctionName(prepareStatement))
		}
	}

	switch strings.ToUpper(operationPtr) {
	case "SELECT":
		stmtOut, err := db.Prepare(fmt.Sprintf("%s %s %s FROM %s %s", operationPtr, flagPtr, columnsPtr, tablePtr, conditionPtr))
		if err != nil {
			log.Error(fmt.Sprintf("%s %s %s FROM %s %s", operationPtr, flagPtr, columnsPtr, tablePtr, conditionPtr))
			log.Error(err.Error())
		}
		rows, err := stmtOut.Query()
		if err != nil {
			log.Error(err.Error())
		}
		defer stmtOut.Close()
		return rows
	case "INSERT":
		stmtOut, err := db.Prepare(fmt.Sprintf("%s %s INTO %s (%s) VALUES (%s)", operationPtr, flagPtr, tablePtr, columnsPtr, conditionPtr))
		if err != nil {
			log.Error(fmt.Sprintf("%s %s INTO %s (%s) VALUES (%s)", operationPtr, flagPtr, tablePtr, columnsPtr, conditionPtr))
			log.Error(err.Error())
		}
		_, err = stmtOut.Exec()
		if err != nil {
			log.Error(err.Error())
		}
		defer stmtOut.Close()
		return nil
	case "DELETE":
		stmtOut, err := db.Prepare(fmt.Sprintf("%s %s FROM %s WHERE %s", operationPtr, flagPtr, tablePtr, conditionPtr))
		if err != nil {
			log.Error(fmt.Sprintf("%s %s FROM %s WHERE %s", operationPtr, flagPtr, tablePtr, conditionPtr))
			log.Error(err.Error())
		}
		_, err = stmtOut.Exec()
		if err != nil {
			log.Error(err.Error())
		}
		defer stmtOut.Close()
		return nil
	case "UPDATE":
		stmtOut, err := db.Prepare(fmt.Sprintf("%s %s %s SET %s", operationPtr, flagPtr, tablePtr, conditionPtr))
		if err != nil {
			log.Error(fmt.Sprintf("%s %s %s SET %s", operationPtr, flagPtr, tablePtr, conditionPtr))
			log.Error(err.Error())
		}
		_, err = stmtOut.Exec()
		if err != nil {
			log.Error(err.Error())
		}
		defer stmtOut.Close()
		return nil
	default:
		log.Error("[%s]: Invalid SQL operation specified. Please check the --operation option.", GetFunctionName(prepareStatement))
	}

	// should not reach here.
	return nil
}

func determineTables(tables string) []string {
	strArr := strings.Split(tables, " ")
	var tablesArr []string
	for i := 0; i < len(strArr); i += 2 {
		tablesArr = append(tablesArr, strArr[i])
	}
	return tablesArr
}

func determineColumns(columns string) []string {
	strArr := strings.FieldsFunc(columns, func(r rune) bool { return r == '.' || r == ',' })
	var colsArr []string
	for i := 1; i < len(strArr); i += 2 {
		colsArr = append(colsArr, strArr[i])
	}
	return colsArr
}

func processData(rows *sql.Rows, columns string, tables string) bool {
	tablesArr := determineTables(tables)
	//columnsArr := determineColumns(columns)

	for rows.Next() {
		switch tablesArr[0] {
		case "employees":
			if len(tablesArr) == 1 { // singular table operation
				err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "salaries" { // JOIN between employees and salaries
				err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate, &salary, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "dept_emp" { // JOIN between employees and dept_emp
				err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate, &deptNo, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}

			} else if tablesArr[1] == "dept_manager" { // JOIN between employees and dept_manager
				err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate, &deptNo, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "titles" { // JOIN between employees and titles
				err := rows.Scan(&empNo, &birthDate, &firstName, &lastName, &gender, &hireDate, &title, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			}
		case "departments":
			if len(tablesArr) == 1 { // singular table operation
				err := rows.Scan(&deptNo, &deptName)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "dept_manager" { // JOIN between departments and dept_manager
				err := rows.Scan(&deptNo, &deptName, &empNo, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "dept_emp" { // JOIN between departments and dept_emp
				err := rows.Scan(&deptNo, &deptName, &empNo, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			}
		case "dept_emp":
			if len(tablesArr) == 1 { // singular table operation
				err := rows.Scan(&empNo, &deptNo, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "employees" { // JOIN between dept_emp and employees
				err := rows.Scan(&empNo, &deptNo, &fromDate, &toDate, &birthDate, &firstName, &lastName, &gender, &hireDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "departments" { // JOIN between dept_emp and departments
				err := rows.Scan(&empNo, &deptNo, &fromDate, &toDate, &deptName)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			}
		case "salaries":
			if len(tablesArr) == 1 { // singular table operation
				err := rows.Scan(&empNo, &salary, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "employees" { // JOIN between salaries and employees
				err := rows.Scan(&empNo, &salary, &fromDate, &toDate, &birthDate, &firstName, &lastName, &gender, &hireDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			}
		case "titles":
			if len(tablesArr) == 1 { // singular table operation
				err := rows.Scan(&empNo, &title, &fromDate, &toDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			} else if tablesArr[1] == "employees" { // JOIN between title and employees
				err := rows.Scan(&empNo, &title, &fromDate, &toDate, &birthDate, &firstName, &lastName, &gender, &hireDate)
				defer rows.Close()
				if err != nil {
					log.Error(err.Error())
				}
			}
		default:
			log.Error("Invalid table specified, please check the --table option.")
			return false
		}
		err := rows.Err()
		defer rows.Close()
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
	fmt.Println(fmt.Sprintf("Running benchmarks, please wait..."))

	sqlQuery := fmt.Sprintf("%s %s %s %s %s", operationPtr, flagPtr, columnsPtr, tablePtr, conditionPtr)
	fmt.Println(fmt.Sprintf("[%s]: Running Query: %s", GetFunctionName(runBenchmarks), sqlQuery))

	if logPrefix == "" {
		log.Warning(fmt.Sprintf("[%s]: No --logprefix defined, log file will NOT be created", GetFunctionName(exportData)))
	}

	br := testing.Benchmark(BenchmarkInitializeDB)
	collectData(br, BenchmarkInitializeDB)

	br = testing.Benchmark(BenchmarkPrepareStatement)
	collectData(br, BenchmarkPrepareStatement)

	br = testing.Benchmark(BenchmarkProcessData)
	collectData(br, BenchmarkProcessData)

	printData()
}

func writeToFile() *os.File {
	filePtr, err := os.OpenFile(fmt.Sprintf("./results/%s.%s.%s", logPrefix, logType, strconv.FormatInt(time.Now().Unix(), 10)), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Error("[%s]: Cannot create file for writing.", GetFunctionName(writeToFile))
	}
	return filePtr
}

func collectData(br testing.BenchmarkResult, funcPtr func(*testing.B)) {
	if logType == "json" && logPrefix != "" {
		benchmarkStr := fmt.Sprintf("%s,%s", br.T, br.N*int(countPtr))
		benchmarkData = append(benchmarkData, string(benchmarkStr))
		benchBuffer = append(benchBuffer, []string{fmt.Sprintf("%s", GetFunctionName(funcPtr)), fmt.Sprintf("%s", br.T), fmt.Sprintf("%d", br.N*int(countPtr)), fmt.Sprintf("%d", br.MemAllocs), fmt.Sprintf("%d", br.MemBytes)})
	} else if logType == "csv" && logPrefix != "" {
		benchmarkStr := fmt.Sprintf("%s,%s", br.T, br.N*int(countPtr))
		benchmarkData = append(benchmarkData, string(benchmarkStr))
		benchBuffer = append(benchBuffer, []string{fmt.Sprintf("%s", GetFunctionName(funcPtr)), fmt.Sprintf("%s", br.T), fmt.Sprintf("%d", br.N*int(countPtr)), fmt.Sprintf("%d", br.MemAllocs), fmt.Sprintf("%d", br.MemBytes)})
	} else {
		benchBuffer = append(benchBuffer, []string{fmt.Sprintf("%s", GetFunctionName(funcPtr)), fmt.Sprintf("%s", br.T), fmt.Sprintf("%d", br.N*int(countPtr)), fmt.Sprintf("%d", br.MemAllocs), fmt.Sprintf("%d", br.MemBytes)})
	}
}

func printData() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Function", "Time Taken", "Iterations", "MemAllocs", "MemBytes"})

	for _, value := range benchBuffer {
		table.Append(value)
	}
	table.Render()
}

func exportData() {
	if logType == "json" {
		filePtr := writeToFile()
		var tempString [][]string
		for _, value := range benchBuffer {
			tempString = append(tempString, value)
		}
		jsonString, _ := json.MarshalIndent(tempString, "", "  ")
		for _, value := range jsonString {
			_, err := filePtr.WriteString(string(value))
			if err != nil {
				log.Error("[%s]: Couldn't write to the JSON output file", GetFunctionName(exportData))
			}
		}
		defer filePtr.Close()
	} else if logType == "csv" {
		filePtr := writeToFile()
		csvWriter := csv.NewWriter(filePtr)
		headerSlice := []string{"function", "timetaken", "iterations", "memallocs", "membytes"}
		_ = csvWriter.Write(headerSlice)
		for _, value := range benchBuffer {
			err := csvWriter.Write(value)
			if err != nil {
				log.Error("[%s]: Cannot write to CSV file", GetFunctionName(exportData))
			}
		}
		csvWriter.Flush()
		defer filePtr.Close()
	} else {
		log.Warning("No --logtype specified, only logging to stdout.")
	}
}

func main() {
	flag.Parse()

	setPtrs()

	validateInput()

	runBenchmarks()

	if logPrefix != "" {
		exportData()
	}
}
