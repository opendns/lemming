#!/bin/bash
# This is a test file to be used when testing the collector.
#
# This test involves use of the Employee Sample Database from Oracle. It is
# availabled at: https://dev.mysql.com/doc/employee/en/
#
# Instructions on how to import this database is at the above address as
# well. Please see them on how to import this database to your test server.

if [[ $1 == '' ]]; then
  max_run=5
else
  max_run=$1
fi

while true; do 
  for i in $(seq 1 $max_run); do
    cat employee-db-tests.sql | mysql -u root employees > /dev/null &
  done
  value=$(( ( RANDOM % 15 )  + 1 ))
  sleep $value
done
