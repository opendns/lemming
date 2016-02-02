MySQL Collector
===============

Overview
--------

MySQL Collector is an application that we use to monitor our MySQL server
infrastructure. This application has two execution paths:

1. Nagios [Default with no options]
2. Graphite

These code paths are explained in depth further in the next section.


Nagios
------

If you would like to execute the collector in Nagios mode do the following:

```

$ ./mysql-collector.py -w 100 -c 500
Threads_connected: 1, Threads_sleeping: 0
$

```

The example above uses the *-w* and the *-t* options which allow the user to
sepcify how many threads connected to the server will return **CRITICAL** or
will return **WARNING**. If you want to try both code paths try using the
following to test both **CRITICAL** and **WARNING**.

```

$ ./mysql-collector.py -w 0 -c 0; echo $?
Threads_connected: 11, Threads_sleeping: 3
2
$ ./mysql-collector.py -w 0 -c 16; echo $?
Threads_connected: 3, Threads_sleeping: 1
1
$ ./mysql-collector.py -w 15 -c 16; echo $?
Threads_connected: 1, Threads_sleeping: 0
0
$

```

Graphite
--------

The graphite option **-G** allows us to dump data into a time series database,
Graphite. This option allows the application to run in *daemon* mode and will
run in an infinate loop until killed with a SIGKILL or SIGTERM signal. This
script can be started by a CRON job.

One can execute this application by running the following (although it is
**HIGHLY** not recommended).

```

$  ./mysql-collector.py -w >> /var/log/mysql-collector.log &

```

Missing?
--------
If there is anything missing from this document one can assume that magic will
be done in the advent that --help is broken.

![alt text](https://media.giphy.com/media/9vTu45DjfiOFa/giphy.gif "MAGIC")
