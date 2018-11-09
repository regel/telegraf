# Radius

The `radius` parser creates metrics from a log file containing radius
accounting events.

### Configuration

```toml
[[inputs.tail]]
  files = ["example"]
  ## Read file from beginning.
  from_beginning = false
  ## Whether file is a named pipe
  pipe = false
  fieldpass = ["Acct-*", "*-Station-Id"]

  ## Method used to watch for file updates.  Can be either "inotify" or "poll".
  # watch_method = "inotify"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "radius"
  ## The format of time data extracted from the first radius event line
  ## this must be specified if `radius_time_key` is specified
  radius_time_format = "Mon Jan 2 15:04:05 2006"
  radius_time_key = "Timestamp"
  ```
#### radius_time_key, radius_time_format

By default the current time will be used for all created metrics, to set the
time using the Radius document you can use the `radius_time_key` and
`radius_time_format` options together to set the time to a value in the parsed
document.

The `radius_time_key` option specifies the column name containing the
time value and `radius_time_format` must be set to a Go "reference time"
which is defined to be the specific time: `Mon Jan 2 15:04:05 MST 2006`.

Consult the Go [time][time parse] package for details and additional examples
on how to set the time format.

### Metrics

One metric is created for each accounting event with the columns added as fields.  The type
of the field is automatically determined based on the contents of the value.

### Examples

Config:
```
[[inputs.tail]]
  files = ["example"]
  from_beginning = false
  pipe = false
  fieldpass = ["Acct-*", "*-Station-Id"]
  data_format = "radius"
  radius_time_format = "Mon Jan 2 15:04:05 2006"
  radius_time_key = "Timestamp"
  radius_measurement_name = "radius"
```

Input:
```
Mon Sep 17 00:02:21 2018
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Start
	Timestamp = 1537135341

Mon Sep 17 00:02:22 2018
	Acct-Session-Id = "00201c14283a008f00841b9ed2ed55ca02c7"
	Acct-Status-Type = Stop
	Timestamp = 1537135342
```

Output:
```
radius Acct-Session-Id=00201c14283a008f00841b9ed2ed55ca02c7,Acct-Status-Type=Start 1537135341000000000
radius Acct-Session-Id=00201c14283a008f00841b9ed2ed55ca02c7,Acct-Status-Type=Stop 1537135342000000000
```
