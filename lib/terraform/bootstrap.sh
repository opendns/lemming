#!/bin/bash

# HACK: Workaround to fix mysql unknown default timezone bug: http://goo.gl/qEES1X
/bin/sed -i -e "s/UTC/\+08\:00/" /etc/my.cnf

# Back up the old data directory
mv /var/lib/mysql/data /var/lib/mysql/data-old

# Define the various parameters for mysqld to know
/opt/software/mysql/scripts/mysql_install_db --datadir=/var/lib/mysql/data --basedir=/opt/software/mysql --defaults-file=/etc/my.cnf






