# MeterReader
A package to count pulses emitted from a LED on a smart meter.
The binary is intended to run at a higher priority so that pulses
are not missed.

Up to 32 different sources can be monitored.

The pulse counts are collated and sent via a UNIX domain socket to
clients that connect locally.
