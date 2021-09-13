# Snowflake

distributed unique ID generator inspired by [Twitter's Snowflake](https://en.wikipedia.org/wiki/Snowflake_ID) with custom bit assignments:

```text
39 bits for time in units of 10 milliseconds(since 2020-01-01 00:00:00 UTC by default), can hold around 174 yrs of time
16 bits for a node ID
 9 bits for a sequence number
```

`snowflake` provide an out of box default settings for usage on public cloud and containers(such as AWS, Azure)
using the lowest 16 bits of the private ip address as node ID.
**NOTE: please don't run multiple instance in same machine/container with the default generator**
