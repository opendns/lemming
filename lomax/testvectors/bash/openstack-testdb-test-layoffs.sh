#!/bin/sh

# Summary: This test case simulates SQL queries in a employee layoff scenario

# Query those employees who make more than > 100,000
./lomax --table="employees a, salaries b" --db="employees" --operation="SELECT" --config=openstack-generic-config.json --cols="a.emp_no, a.birth_date, a.first_name, a.last_name, a.gender, a.hire_date, b.salary, b.from_date, b.to_date" --condition="where b.salary > 100000 limit 10" --user="root" --password="password"

# TODO: Add more queries for this scenario 
