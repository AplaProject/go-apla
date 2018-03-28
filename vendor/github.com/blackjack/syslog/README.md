syslog
======

Golang alternative for built-in log/syslog package

Differences with standard library
--------

The "log/syslog" package writes raw data directly into syslog socket.
The resulting log message will look like: 

    Sep 18 10:28:52 server 2013-09-18T10:28:52Z server [programname][20310]: Log message 

You see that server name and datetime are duplicated. The log message format is hardcoded into source, 
thus you cannot omit data duplication. 

Another major disadvantage is that resulting log message cannot be processed using tools such as rsyslog, 
because it cannot obtain the name of the sender application.

The approach used in this library is calling C functions openlog and syslog directly.
It solves both problems mentioned above.

The library provides:
* Openlog, Syslog functions with parameters identical to those in syslog.h C header
* log/syslog-like functions for writing messages: Emerg, Alert, Crit etc.
* Their formatted versions: Emergf, Alertf, Critf etc.
* io.Writer interface


Restrictions
--------

Because of using C-strings for calling syslog functions, to omit memory leaks 
when you call Openlog function the data from **ident** parameter is copied into 
static char array with fixed length. Length of this array is defined in syslog_wrapper.h
and currently equals **1000**. If length if ident string exceeds it, then 
it will be cut up to maximum value.

Example
--------

    import "github.com/blackjack/syslog"
    func main() {
        syslog.Openlog("awesome_app", syslog.LOG_PID, syslog.LOG_USER)
        syslog.Syslog(syslog.LOG_INFO, "Hello syslog!")
        syslog.Err("Sample error message")
        syslog.Critf("Sample %s crit message", "formatted")
    }


The resulting log message will look like: 

    Sep 18 18:13:46 server awesome_app[16844]: Hello syslog!
    Sep 18 18:13:46 server awesome_app[16844]: Sample error message
    Sep 18 18:13:46 server awesome_app[16844]: Sample formatted crit message

Feel the difference!
