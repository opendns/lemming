#!/bin/sh

# Summary: This test case simulates SQL queries in a employee paycut scenario

# Preparation Steps
go build lomax.go

# Query those employees who make more than > 100,000
./lomax --db="employees" --operation="SELECT" --table="employees a, salaries b" --condition="WHERE  b.salary > 100000 limit 1" --cols="a.emp_no, a.birth_date, a.first_name, a.last_name, a.gender, a.hire_date, b.salary, b.from_date, b.to_date" --user="root" --password="password" --config=openstack-generic-config.json

# Insert a Random employee
./lomax --db="employees" --operation="INSERT" --flag="IGNORE" --table="employees" --random="true" --config=openstack-generic-config.json --user="root" --password="password" --count=1

# Insert a Random employee
./lomax --db="employees" --operation="INSERT" --flag="IGNORE" --table="employees" --cols="emp_no, birth_date, first_name, last_name, gender, hire_date" --condition="1010101, 1980-01-01, 'John', 'Doe', 'M', 2016-01-01" --config=openstack-generic-config.json --user="root" --password="password" --count=1

# Update the employees with > 100,000 salary from employees table 
./lomax --db="employees" --operation="UPDATE" --table="salaries" --condition="salary=90000 WHERE salary > 100000 limit 1" --cols=" " --config=openstack-generic-config.json --user="root" --password="password" --count=1

# Delete the employees with 100,000 salary from dept_emp table
./lomax --db="employees" --operation="DELETE" --table="dept_emp" --condition="emp_no IN ( SELECT emp_no FROM ( SELECT salary FROM salaries WHERE salary > 100000) as p) limit 1" --config=openstack-generic-config.json --user="root" --password="password"

# Delete the employees with 100,000 salary from dept_manager table
./lomax --db="employees" --operation="DELETE" --table="dept_manager" --condition="emp_no IN ( SELECT emp_no FROM ( SELECT salary FROM salaries WHERE salary > 100000) as p) limit 1" --config=openstack-generic-config.json --user="root" --password="password"

# Delete the employees with 100,000 salary from titles table
./lomax --db="employees" --operation="DELETE" --table="titles" --condition="emp_no IN ( SELECT emp_no FROM ( SELECT salary FROM salaries WHERE salary > 100000) as p) limit 1" --config=openstack-generic-config.json --user="root" --password="password"

# Delete the employees with 100,000 salary from salary table
./lomax --db="employees" --operation="DELETE" --table="salaries" --condition="salary > 100000 limit 1" --config=openstack-generic-config.json --user="root" --password="password"
