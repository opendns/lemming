#!/bin/sh

# Summary: This test case simulates SQL queries in a employee hiring spree. 

# Preparation Steps
go build lomax.go

# Add newly hired employees to the employees database
./lomax --db="employees" --operation="INSERT" --flag="IGNORE" --table="employees" --random="true" --config=openstack-generic-config.json --user="root" --password="password" --count=10000


