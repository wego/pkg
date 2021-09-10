# Snowflake

distributed unique ID generator inspired by [Twitter's Snowflake](https://en.wikipedia.org/wiki/Snowflake_ID) with custom bit assignments:

```text
39 bits for time in units of 10 msec
 8 bits for a sequence number
16 bits for a node ID
```

`snowflake` provide an out of box default settings for usage on public cloud and containers(such as AWS, Azure)
using the lowest 16 bits of the private ip address as node ID.
**NOTE: please don't run multiple instance in same machine/container with the default generator**
