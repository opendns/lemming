# killtracer

`killtracer` is a program for monitoring when signals are sent to processes on
a given host.  It will print a log message whenever something interesting is seen.
It requires Linux system call tracing to be compiled and enabled.

## Requirements
- Linux system call tracing (a kernel debug feature) must be compiled and enabled.
- `killtracer` must run as root.

## Example output

<pre>
2016-01-08 22:01:59.218831 [INFO]: Signal detected: target PID: 2514; signal: 28; exit 0x0; source process: bash-1958; source UID: 0; source EUID: 0
2016-01-08 22:02:11.126948 [INFO]: Signal detected: target PID: 2514; signal: 28; exit 0x7FFFFFFFFFFFFFFF; source process: sh-32576; source UID: 12345; source EUID: 12345 
</pre>

Notice the second system call has a non-zero exit value, indicating it did not succeed.

Signal numbers can be found in the signal(7) man page on most Linux distros.

## Example init.d Start Script
Starting at Boot time on a Debian based systes may be useful if you want to include long term logging of signals. Do the following to add `killtracer` to your startup and shutdown initialization processes.

<pre>
# sudo cp killtracer/killtracer.init /etc/init.d/killtracer
# sudo update-rc.d killtracer defaults 98 02
</pre>
