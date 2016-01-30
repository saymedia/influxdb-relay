influxdb-relay
==============

`influxdb-relay` is a small proxy server that is intended to act as a local
InfluxDB UDP listener, collect UDP messages, and then relay those messages on
to a remote InfluxDB server via the HTTP protocol.

This local relay allows you to "fire and forget" InfluxDB messages into a
local buffer, and then have them be reasonably likely to be delivered to your
InfluxDB server.

This tool may be useful to you, with the following assumptions:

* You don't mind losing data under load or network problems: by design, this
  tool has a fixed-size buffer and will drop messages if that buffer gets full.

* You want to isolate your application from InfluxDB uptime: if you're
  writing supporting metrics to InfluxDB in your main request path, you'd like
  to collect as much data as possible but you don't want to degrade your app
  if InfluxDB is down or slow, and you don't want your application to be
  complicated by having to gracefully recover from InfluxDB errors.

The design of `influxdb-relay` is very simple: it allocates a bunch of memory
to use as a buffer, and then waits for packets to arrive on its UDP port. When
packets show up, they are read into the buffer and placed into a queue to
write to the InfluxDB backend server. If UDP packets arrive faster than they
can be transmitted to InfluxDB then the buffer will eventually fill, at which
time incoming packets will initially be placed into the OS socket buffer, after
which they will be dropped altogether.

Thus it is important to configure the relay with the appropriate size of buffer
to absorb transmission delays to the InfluxDB server, *and* to ensure that
the connection to your InfluxDB server is fast and reliable enough so that
the queue can keep up.

If the InfluxDB server is temporarily unavailable then the relay queue will
start to fill, and once full new data will be dropped. However, the relay
should automatically recover when InfluxDB returns to service, depleting
whatever is in the queue and then accepting new packets.

Configuration
-------------

```
Usage of influxdb-relay:
  -buffer-size=4096: Maximum number of packets that can be buffered
  -listen-addr="127.0.0.1:4444": Local address for the UDP listener
  -max-line-length=256: Maximum line length for line protocol, in bytes
  -target-url="http://127.0.0.1:8086/write?db=example": URL where recieved data should be written
```

License
-------

The MIT License (MIT)

Copyright (c) 2016 Say Media Inc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
