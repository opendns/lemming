#!/bin/sh

counter="0"

if [ -z $1 ]; then
    echo "Usage: ./stress_test-1.sh <spawnCount>"
    echo "Please input a spawn count as the first argument"
    exit 1
fi

echo "Running stress test 1 for $1 times in the background..."

while [ $counter -lt $1 ]
do
./testvectors/bash/openstack-testdb-test-company-expansion.sh > /dev/null 2>&1 &
./testvectors/bash/openstack-testdb-test-new-hires.sh > /dev/null 2>&1 &
./testvectors/bash/openstack-testdb-test-paycuts-and-layoffs.sh > /dev/null  2>&1 &
counter=$[$counter+1]
done
