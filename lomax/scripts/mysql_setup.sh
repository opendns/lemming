#!/bin/sh

# Clone and setup the test_db repo
git clone https://github.com/datacharmer/test_db.git
mysql < test_db/employees.sql
