package main

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestTravis(t *testing.T) {
	configParse("openstack-generic-config.json", "openstack-generic-tests.json")
	cases := []struct {
		testTableInput     string
		testHostNameInput  string
		testPortInput      string
		testDBInput        string
		testOperationInput string
		testFlagInput      string
		testRandomInput    string
		testColumnsInput   string
		testConditionInput string
		outputWanted       bool
	}{
		{"employees", "localhost", "3306", "employees", "SELECT", " ", "false", "*", "LIMIT 2", true},
	}
	for _, c := range cases {
		db := initializeDB(USER, PASSWORD, c.testHostNameInput, c.testPortInput, c.testDBInput)
		returnedRows := prepareStatement(db, c.testOperationInput, c.testFlagInput, c.testRandomInput, c.testColumnsInput, c.testTableInput, c.testConditionInput)
		returnedOutput := processData(returnedRows, c.testColumnsInput, c.testTableInput)
		if returnedOutput != c.outputWanted {
			t.Errorf("prepareStatement(%q, %q, %q, %q, %q) returned %q, want %q", c.testDBInput, c.testColumnsInput, c.testOperationInput, c.testFlagInput, c.testRandomInput, c.testTableInput, c.testConditionInput, returnedOutput, c.outputWanted)
		}
	}
}
