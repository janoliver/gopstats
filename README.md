# Gopstats: A milter that collects statistics

Gopstats is a simple `milter` (mail filter) intended to collect some statistics of a mail server. 
At the moment, it simply writes some information about each mail that it is notified of to an SQLite database.
It is up to the user to write clients that present the collected data in a good way. 

Gopstats was written to supersede syslog parsers, which are not really real-time. Due to the limitations that
come with the milter concept, it cannot collect responses of the receiving mail server.

## installation

Build the program and copy the `gopstats` binary somewhere onto the mail server. Change and use `gopstats.service` for 
systemd, or run the binary by hand. Command line parameters are `-port` to change the port and `-db-path` to change the
path to the SQLite database.

### Postfix integration

Add the following entry to `smtp_milters` setting: `inet:localhost:9929`. For example, the line could look like that:

    smtpd_milters = some_other_milters, inet:localhost:9929
    
