# Server
A package to connect to a MeterReader instance, gather the
accumulated data, and allow it to be retrieved via a web server.
The server of the data is separate from the reading to allow the reader to
run at a higher priority.

The URL can include an 'after' parameter, a Unix time so that only values
time stamped after this time are returned.
