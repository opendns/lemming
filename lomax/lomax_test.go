package main

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestTravis(t *testing.T) {
	cases := []struct {
		testTableInput     string
		testDBInput        string
		testOperationInput string
		testConditionInput string
		outputWanted       bool
	}{
		{"departments", "employees", "SELECT *", "", true},
	}
	for _, c := range cases {
		db := initializeDB(USER, PASSWORD, c.testDBInput)
		returnedRows := prepareStatement(db, c.testOperationInput, c.testTableInput, c.testConditionInput)
		returnedOutput := processData(returnedRows)
		if returnedOutput != c.outputWanted {
			t.Errorf("prepareStatement(%q, %q, %q, %q) returned %q, want %q", c.testDBInput, c.testOperationInput, c.testTableInput, c.testConditionInput, returnedOutput, c.outputWanted)
		}
	}
}
