# Stackdriver Monitoring Exporter

Export metric points from Stackdriver Monitoring to csv files.

Using go version 1.11 or above.

## Setup Credentials

Set the creditails permission first, you may create a service account with Role *Monitoring Viewer*

Use current login permission:

```shell
$ gcloud auth application-default login
```

Use the Service Account's crendentials

```shell
$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
```

Edit the `config.yaml`

```shell
$ cp config.yaml.example config.yaml
```

Example:

```yaml
# https://cloud.google.com/monitoring/api/metrics_gcp
projects:
  - projectID: <YOUR_PROJECT_ID>
    metrics:
    - title: vm-cpu-usage_time
      metricType: compute.googleapis.com/instance/cpu/usage_time
      andFilters:
        - metric.labels.instance_name="gce-instance-name"
      intervalType: day
      timezone: 8
      unit: s
      aggregationPerSeriesAligner: ALIGN_RATE
      aggregationAlignmentPeriod: 300s
    - title: vm-cpu-network-sent_bytes_count
      metricType: compute.googleapis.com/instance/network/sent_bytes_count
      andFilters:
        - metric.labels.instance_name="gce-instance-name"
      intervalType: day
      timezone: 8
      unit: bytes
      aggregationPerSeriesAligner: ALIGN_RATE
      aggregationAlignmentPeriod: 300s
```

Replace `<YOUR_PROJECT_ID>` with your project ID, and `gce-instance-name` to your instance name.

Execute to download the metric points:

```shell
$ go run main.go
```

The metrics csv will be generated to metrics directory.

```shell
metrics/
├── 2018-10-18T00:00:00[vm-cpu-usage_time][s].csv
└── 2018-10-18T00:00:00[vm-cpu-network-sent_bytes_count][bytes].csv
```

Content is like:

```plain
timestamp,datetime,value
1539821100,2018-10-18 00:05:00,0.024325785464607178
1539821400,2018-10-18 00:10:00,0.019803228156330684
1539821700,2018-10-18 00:15:00,0.027172103356181955
1539822000,2018-10-18 00:20:00,0.020661692447029055
1539822300,2018-10-18 00:25:00,0.025935437206644565
1539822600,2018-10-18 00:30:00,0.023403971735776092
...
```