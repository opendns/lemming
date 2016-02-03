#!/usr/bin/env python

import os
import re
import sys
import pwd
import grp
import json
import argparse
import socket
import MySQLdb as mdb
from time import time, sleep
from pprint import pprint

DEBUG = False

class QPSData():
    def __init__(self, file_path='/var/tmp/mysql-collector.qps.txt'):
        self.unixtime  = time()
        self.delete    = 0
        self.insert    = 0
        self.update    = 0
        self.select    = 0
        self.total_qps = 0
        self.file_path = file_path

        # Grab UID and PID so we can lock down the stats file.
        self.uid  = pwd.getpwnam("root").pw_uid
        self.gid  = grp.getgrnam("root").gr_gid

        # Check to see if the data file exists. If it does we should read it and
        # load the data into this class object, otherwise we should create a
        # blank file that contains the 0s above.
        if os.path.exists(file_path):
            os.chown(self.file_path, self.uid, self.gid)    # Set to root owner
            os.chmod(self.file_path, 384)                   # 0600 in octal
            self.__dict__ = json.loads(self.__parse_file__(file_path))
        else:
            self.update_file()

    def update_file(self):
        with open(self.file_path, 'wb+') as file:
            json.dump(self.__dict__, file)
            file.write("\n")
        os.chown(self.file_path, self.uid, self.gid)    # Set to root owner
        os.chmod(self.file_path, 384)                   # 0600 in octal

    # We need to check to see if the temp calculations file exists, if it is
    # missing we should create a blank one.
    @classmethod
    def __parse_file__(cls, file_path):
        data = None
        with open(file_path, "r") as file:
            data = file.readline().rstrip('\n')
        return data


def parseArgs():
    global DEBUG
    parser = argparse.ArgumentParser(description="mysql-collector is a tool for MySQL stats collection.")

    parser.add_argument("-u",
                        "--user",
                        dest="user",
                        metavar="user",
                        default="user",
                        help="The MySQL username to login as.")
    parser.add_argument("-p",
                        "--password",
                        dest="password",
                        metavar="password",
                        default="password",
                        help="The MySQL password to login as.")
    parser.add_argument("-c",
                        "--critical-threshold",
                        type=int,
                        dest="critical_threshold",
                        metavar="500",
                        default=500,
                        help="Threads threshold for Nagios level CRITICAL.")
    parser.add_argument("-d",
                        "--debug",
                        dest="debug",
                        default=False,
                        action="store_true",
                        help="Enables Debug mode.")
    parser.add_argument("-f",
                        "--threads-file",
                        dest="threads_file",
                        default="/var/log/mysql-processlist.log",
                        action="store_true",
                        help="Store metric in graphite.")
    parser.add_argument("-w",
                        "--warning-threshold",
                        type=int,
                        dest="warning_threshold",
                        metavar="400",
                        default=400,
                        help="Threads threshold for Nagios level WARNING")
    parser.add_argument("-t",
                        "--threads-logger",
                        dest="threads_logger",
                        default=False,
                        action="store_true",
                        help=("Enable SHOW PROCESSLIST if CRITICAL threshold"
                              "is hit."))
    parser.add_argument("-S",
                        "--graphite-server",
                        dest="graphite_server",
                        metavar="graphite_server",
                        default="127.0.0.1",
                        help="Graphite hostname for storing metrics.")
    parser.add_argument("-G",
                        "--graphite",
                        dest="graphite",
                        default=False,
                        action="store_true",
                        help="Store metric in graphite.")
    parser.add_argument("-I",
                        "--ignore-lock",
                        dest="ignore_lock",
                        default=False,
                        action="store_true",
                        help=("Ignore the lock file created by this script. "
                              "Don't use this unless testing, because you risk"
                              "breaking replication and dropping tables"
                              "randomly."))

    args = parser.parse_args()
    if args.debug:
        DEBUG = True
    return args

def lock_file(pid_file='/var/run/mysql-collector.pid'):
    old_pid = None
    this_pid = int(os.getpid())
    if os.path.exists(pid_file):
        with open(pid_file, 'r') as file:
            old_pid = int(file.readline())
        pids = [int(p) for p in os.listdir('/proc') if p.isdigit()]
        if old_pid in pids:
            sys.exit(0)
        else:
            os.remove(pid_file)
    with open(pid_file, 'wb+') as file:
        file.write("%d\n" % (this_pid))

def mysql_query(sql,
                args,
                dict_cursor=False):
    result = None
    conn = None
    cur = None
    try:
        conn = mdb.connect(args.graphite_server, args.user, args.password)
        if dict_cursor:
            cur = conn.cursor(cursorclass=mdb.cursors.DictCursor)
        else:
            cur = conn.cursor()
        cur.execute(sql)
        result = cur.fetchall()
    except Exception as e:
        print("Error %s" % (e))
        sys.exit(2)
    finally:
        if conn:
            conn.close()
    return result


def get_qps(qps_data):
    calculated_data = {}
    # This is the list that we will use to build our QPS data from.
    # These should be valid status items in MySQL like the SQL string below.
    item_list = [
        'Com_delete',
        'Com_insert',
        'Com_update',
        'Com_select',
        'Questions'
    ]
    sql_string = ("','".join(item_list))
    sql = ("SHOW GLOBAL STATUS WHERE Variable_name IN ('%s')" % sql_string)
    results = mysql_query(sql, dict_cursor=True)
    unixtime = int(time())
    time_difference = abs(unixtime - qps_data.unixtime)
    qps_data.unixtime = unixtime
    for item in results:
        key = item['Variable_name']
        value = int(item['Value'])
        # print("%d\t\t%s: %d" % (time_difference, key, value))
        if 'Com_delete' in key:
            calculated_data['delete'] = (abs(value - qps_data.delete) /
                                             time_difference)
            qps_data.delete = value

        elif 'Com_insert' in key:
            calculated_data['insert'] = (abs(value - qps_data.insert) /
                                             time_difference)
            qps_data.insert = value

        elif 'Com_update' in key:
            calculated_data['update'] = (abs(value - qps_data.update) /
                                             time_difference)
            qps_data.update = value

        elif 'Com_select' in key:
            calculated_data['select'] = (abs(value - qps_data.select) /
                                             time_difference)
            qps_data.select = value

        elif 'Questions' in key:
            calculated_data['total_qps'] = (abs(value - qps_data.total_qps) /
                                                time_difference)
            qps_data.total_qps = value
        else:
            continue

    # Save data to disk in-case we reboot or crash.
    qps_data.update_file()
    return calculated_data

def store_metric(key, value, graphite_server):
    # Debug assumes we do not have access to the Graphite cluster.
    if not DEBUG:
        hostname = os.uname()[1]
        hostname_split = hostname.split('.')
    else:
        hostname_split = ('DEBUG', os.uname()[0])

    if not DEBUG:
        sock = socket.socket()
        sock.connect((graphite_server, 2003))
    message = ("mysql.%s.%s.%s %s %d\n" % (hostname_split[1],
                                           hostname_split[0],
                                           key,
                                           value,
                                           int(time())))
    if not DEBUG:
        sock.sendall(message)
        if sock:
            sock.close()
    #print message

def threads_sleeping(args):
    threads = 0
    sql = "SHOW FULL PROCESSLIST"
    results = mysql_query(sql, args, dict_cursor=True)
    for row in results:
        if 'Command' in row:
            if row['Command'] in 'Sleep':
                threads += 1
    return threads

def graphite_run(args):
    qps_data = QPSData()

    # Verify that the script is already not running
    if not args.ignore_lock and not DEBUG:
        lock_file()

    # Run forever
    while True:
        connections = 0
        # Grab mysql status
        collection_group = ['Threads_connected',
                            'Created_tmp_disk_tables',
                            'Handler_read_first',
                            'Innodb_buffer_pool_wait_free',
                            'Key_reads',
                            'Max_used_connections',
                            'Open_tables',
                            'Select_full_join',
                            'Slow_queries',
                            'Uptime']

        # Grab innodb buffer pool stats
        sql = ("SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool%%'")
        bp_results = mysql_query(sql)
        for key, value in bp_results:
            if type(value) == 'int':
                store_metric(key, int(value), args.graphite_server)

        # Grab connections from local MySQL server
        sql = "SHOW STATUS"
        tc_results = mysql_query(sql)
        for key, value in tc_results:
            if key in collection_group:
                store_metric(key, value, args.graphite_server)
            if key == 'Threads_connected':
                connections = int(value)

        # Grab the number of threads that are asleep
        threads_asleep = threads_sleeping()
        store_metric('Threads_sleeping', threads_asleep, args.graphite_server)

        # Now write QPS to graphite
        qps_dict = get_qps(qps_data)
        for item in qps_dict:
            store_metric(item, qps_dict[item], args.graphite_server)

        # Grab slave status
        try:
            results = mysql_query("SHOW SLAVE STATUS")
            match = re.search(r'Seconds_Behind_Master:\s(\d+)', results)
            if match:
                sbm = int(match.group(1))
            else:
                sbm = 0
            store_metric('Seconds_Behind_Master', sbm, args.graphite_server)
        except:
            pass
        if DEBUG:
            pprint(locals())
        sleep(10)

def main(options):
    args = options
    # If the graphite option is passed in we should skip the Nagios stuff and
    # vise versa.
    if args.graphite:
        graphite_run(args)
        # Since graphite run is a while true loop we should never get here.
        sys.exit(1)

    # BELOW IS FOR NAGIOS ONLY
    # Grab number of running threads for Nagios alarm
    results = mysql_query("SHOW STATUS", args)
    for key, value in results:
        if key == 'Threads_connected':
            connections = int(value)

    threads_asleep = threads_sleeping(args)
    store_metric('Threads_sleeping', threads_asleep, args.graphite_server)
    print("Threads_connected: %d, Threads_sleeping: %d" % (connections,
                                                           threads_asleep))
    if (connections < args.critical_threshold and
        connections >= args.warning_threshold):
        # NAGIOS: WARNING
        return 1
    elif connections >= args.critical_threshold:
        # NAGIOS: CRITICAL
        return 2
    else:
        # NAGIOS: OK
        return 0

    # Should never get here
    # NAGIOS: UNKNOWN
    return 3

if __name__ == "__main__":
    val = main(parseArgs())
    sys.exit(val)
