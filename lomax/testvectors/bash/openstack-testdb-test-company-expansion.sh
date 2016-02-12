#!/bin/sh

# Summary: This test case simulates SQL queries in a company expansion. 

# Preparation Steps
go build lomax.go

# Add new departments to our company
./lomax --db="employees" --operation="INSERT" --flag="IGNORE" --table="departments" --random="true" --config=openstack-generic-config.json --user="root" --password="password" --count=1000 --logtype=csv --logprefix=test-company-expansion

# Add newly hired employees to the employees database
./lomax --db="employees" --operation="INSERT" --flag="IGNORE" --table="employees" --random="true" --config=openstack-generic-config.json --user="root" --password="password" --count=1000  --logtype=csv --logprefix=test-company-expansion

# Give low earning employees a 10% salary bonus
./lomax --db="employees" --operation="UPDATE" --table="salaries" --condition="salary=11000 WHERE salary < 10000" --cols=" " --config=openstack-generic-config.json --user="root" --password="password"  --logtype=csv --logprefix=test-company-expansion

# Call a meeting with the top earning employees
./lomax --db="employees" --operation="SELECT" --table="employees a, salaries b" --condition="WHERE  b.salary > 100000" --cols="a.emp_no, a.birth_date, a.first_name, a.last_name, a.gender, a.hire_date, b.salary, b.from_date, b.to_date" --user="root" --password="password" --config=openstack-generic-config.json  --logtype=csv --logprefix=test-company-expansion
