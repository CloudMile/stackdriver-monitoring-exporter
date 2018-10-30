# Stackdriver Monitoring Exporter

Export metric points from Stackdriver Monitoring to csv files.

Using go version 1.11 or above.

## Enable CLoud API

```shell
$ gcloud services enable cloudresourcemanager.googleapis.com
$ gcloud services enable monitoring.googleapis.com
```

## Configuration

Edit the `config.yaml`

```shell
$ cp config.yaml.example config.yaml
```

Example:

```yaml
# https://cloud.google.com/monitoring/api/metrics_gcp
exporter: GCSExporter
destination: <GCS_BUCKET_NAME>
```

GCSExporter'destination is Google Cloud Storage Bucket Name. The service acccount has to be grant the **Storage Object Admin** permission of Bucket.

## Development

```shell
$ dev_appserver.py app.yaml
```

## Deploymenty

```shell
$ gcloud app deploy app.yaml cron.yaml
```

## Export

The metrics csv will be generated to metrics directory.

```shell
<destination>/
└── <project_id>
    └── 2018
        └── 10
            └── 18
                └── instance_name
                    ├── 2018-10-18T00:00:00[instance_name][cpu_usage_time].csv
                    ├── 2018-10-18T00:00:00[instance_name][network_received_bytes_count].csv
                    └── 2018-10-18T00:00:00[instance_name][network_sent_bytes_count].csv
```

File content looks like:

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