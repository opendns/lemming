#!/bin/sh

# Clone and setup the test_db repo
git clone https://github.com/datacharmer/test_db.git
cd ./test_db
mysql < ./employees.sql
