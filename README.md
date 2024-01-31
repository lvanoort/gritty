### Warning: this is a personal experiment with something I find interesting. It is not intended to be an example of good code or good design.

Gritty is an experiment in writing a trace-based testing library for Go. Unlike many other
trace-based testing libraries, it is intended to be a seamless(ish) integration into standard
Go unit tests rather than running separately.