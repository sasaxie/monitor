# Monitor

## A Monitor for java-tron

Monitor is a open source monitor for java-tron. It's useful for monitoring nodes
 of java-tron.

## Features

- Monitor NowBlockNum, gRPC's ping, LastSolidityBlockNum.

## Getting Started

### Install InfluxDB

1. Install InfluxDB

```shell
docker pull influxdb
```

2. Modify influxdb.conf

```shell
cd influxdb/
docker run --rm influxdb influxd config > influxdb.conf
sed "s/auth-enabled = false/auth-enabled = true/g" influxdb.conf > tmp
cat tmp > influxdb.conf
```

3. Run InfluxDB

```shell
docker run -p 8086:8086 -e INFLUXDB_ADMIN_USER=tron -e INFLUXDB_ADMIN_PASSWORD=trondb --name influxdb_monitor -v $PWD/influxdb.conf:/etc/influxdb/influxdb.conf influxdb -config /etc/influxdb/influxdb.conf
```

4. Create database

```shell
curl -XPOST http://localhost:8086/query\?u=tron\&p=trondb --data-urlencode "q=CREATE DATABASE tronmonitor"
```

5. Tips

You can stop influxdb by command `docker stop influxdb_monitor`.
You can start influxdb by command `docker start influxdb_monitor`.

### Install Monitor

1. Install Monitor

```shell
cd monitor/
docker pull sasaxie/tron-monitor
```

2. Modify monitor.toml

```shell
docker run sasaxie/tron-monitor cat /go/src/github.com/sasaxie/monitor/conf/monitor.toml > monitor.toml
```

3. Modify nodes.json

```shell
docker run sasaxie/tron-monitor cat /go/src/github.com/sasaxie/monitor/conf/nodes.json > nodes.json
```

4. Run Monitor

```shell
docker run --link influxdb_monitor --name tron-monitor -v $PWD/monitor.toml:/go/src/github.com/sasaxie/monitor/conf/monitor.toml -v $PWD/nodes.json:/go/src/github.com/sasaxie/monitor/conf/nodes.json sasaxie/tron-monitor
```

5. Tips

You can stop monitor by command `docker stop tron-monitor`.
You can start monitor by command `docker start tron-monitor`.

### Install Grafana

1. Install Grafana

```shell
docker run -p 3000:3000 --name grafana_monitor grafana/grafana
```

Open grafana in your browser (default: http://localhost:3000) and login with admin user (default: user/pass = admin/admin).

2. Add Data Source

Name: influxdb
Type: InfluxDB

HTTP URL: http://localhost:8086
HTTP Access: Browser

InfluxDB Details Database: tronmonitor
InfluxDB Details User: tron
InfluxDB Details Password: trondb

[Save & Test]

3. Import DashBoard

[Create] - [Import]

[Or paste JSON]

```json
{
  "annotations": {
    "list": [
      {
        "builtIn": 1,
        "datasource": "-- Grafana --",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations & Alerts",
        "type": "dashboard"
      }
    ]
  },
  "editable": true,
  "gnetId": null,
  "graphTooltip": 0,
  "id": 6,
  "links": [],
  "panels": [
    {
      "alert": {
        "conditions": [
          {
            "evaluator": {
              "params": [
                1
              ],
              "type": "lt"
            },
            "operator": {
              "type": "and"
            },
            "query": {
              "params": [
                "A",
                "5m",
                "now"
              ]
            },
            "reducer": {
              "params": [],
              "type": "avg"
            },
            "type": "query"
          }
        ],
        "executionErrorState": "alerting",
        "frequency": "40s",
        "handler": 1,
        "message": "54.236.37.243 ping timeout",
        "name": "54.236.37.243 ping timeout",
        "noDataState": "no_data",
        "notifications": []
      },
      "aliasColors": {},
      "bars": false,
      "dashLength": 10,
      "dashes": false,
      "datasource": "InfluxDB",
      "fill": 1,
      "gridPos": {
        "h": 16,
        "w": 18,
        "x": 0,
        "y": 0
      },
      "id": 2,
      "legend": {
        "avg": false,
        "current": true,
        "max": true,
        "min": true,
        "show": true,
        "total": false,
        "values": true
      },
      "lines": true,
      "linewidth": 1,
      "links": [],
      "nullPointMode": "null",
      "percentage": false,
      "pointradius": 5,
      "points": false,
      "renderer": "flot",
      "seriesOverrides": [],
      "spaceLength": 10,
      "stack": false,
      "steppedLine": false,
      "targets": [
        {
          "alias": "ping",
          "groupBy": [
            {
              "params": [
                "$__interval"
              ],
              "type": "time"
            },
            {
              "params": [
                "previous"
              ],
              "type": "fill"
            }
          ],
          "measurement": "node_status",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "ping"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "mean"
              }
            ]
          ],
          "tags": [
            {
              "key": "node",
              "operator": "=",
              "value": "54.236.37.243"
            }
          ]
        }
      ],
      "thresholds": [
        {
          "colorMode": "critical",
          "fill": true,
          "line": true,
          "op": "lt",
          "value": 1
        }
      ],
      "timeFrom": null,
      "timeShift": null,
      "title": "Ping",
      "tooltip": {
        "shared": true,
        "sort": 0,
        "value_type": "individual"
      },
      "type": "graph",
      "xaxis": {
        "buckets": null,
        "mode": "time",
        "name": null,
        "show": true,
        "values": []
      },
      "yaxes": [
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        },
        {
          "format": "short",
          "label": null,
          "logBase": 1,
          "max": null,
          "min": null,
          "show": true
        }
      ],
      "yaxis": {
        "align": false,
        "alignLevel": null
      }
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": true,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": null,
      "format": "none",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 5,
        "w": 6,
        "x": 18,
        "y": 0
      },
      "id": 4,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": false,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "groupBy": [
            {
              "params": [
                "$__interval"
              ],
              "type": "time"
            },
            {
              "params": [
                "null"
              ],
              "type": "fill"
            }
          ],
          "measurement": "node_status",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "NowBlockNum"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "last"
              }
            ]
          ],
          "tags": [
            {
              "key": "node",
              "operator": "=",
              "value": "54.236.37.243"
            }
          ]
        }
      ],
      "thresholds": "",
      "title": "Now Block Num",
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    },
    {
      "cacheTimeout": null,
      "colorBackground": false,
      "colorValue": true,
      "colors": [
        "#299c46",
        "rgba(237, 129, 40, 0.89)",
        "#d44a3a"
      ],
      "datasource": null,
      "format": "none",
      "gauge": {
        "maxValue": 100,
        "minValue": 0,
        "show": false,
        "thresholdLabels": false,
        "thresholdMarkers": true
      },
      "gridPos": {
        "h": 5,
        "w": 6,
        "x": 18,
        "y": 5
      },
      "id": 6,
      "interval": null,
      "links": [],
      "mappingType": 1,
      "mappingTypes": [
        {
          "name": "value to text",
          "value": 1
        },
        {
          "name": "range to text",
          "value": 2
        }
      ],
      "maxDataPoints": 100,
      "nullPointMode": "connected",
      "nullText": null,
      "postfix": "",
      "postfixFontSize": "50%",
      "prefix": "",
      "prefixFontSize": "50%",
      "rangeMaps": [
        {
          "from": "null",
          "text": "N/A",
          "to": "null"
        }
      ],
      "sparkline": {
        "fillColor": "rgba(31, 118, 189, 0.18)",
        "full": false,
        "lineColor": "rgb(31, 120, 193)",
        "show": true
      },
      "tableColumn": "",
      "targets": [
        {
          "groupBy": [
            {
              "params": [
                "$__interval"
              ],
              "type": "time"
            },
            {
              "params": [
                "null"
              ],
              "type": "fill"
            }
          ],
          "measurement": "node_status",
          "orderByTime": "ASC",
          "policy": "default",
          "refId": "A",
          "resultFormat": "time_series",
          "select": [
            [
              {
                "params": [
                  "LastSolidityBlockNum"
                ],
                "type": "field"
              },
              {
                "params": [],
                "type": "last"
              }
            ]
          ],
          "tags": [
            {
              "key": "node",
              "operator": "=",
              "value": "54.236.37.243"
            }
          ]
        }
      ],
      "thresholds": "",
      "title": "Last Solidity Block Num",
      "type": "singlestat",
      "valueFontSize": "80%",
      "valueMaps": [
        {
          "op": "=",
          "text": "N/A",
          "value": "null"
        }
      ],
      "valueName": "current"
    }
  ],
  "refresh": false,
  "schemaVersion": 16,
  "style": "dark",
  "tags": [
    "FullNode",
    "MainNet"
  ],
  "templating": {
    "list": []
  },
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {
    "refresh_intervals": [
      "5s",
      "10s",
      "30s",
      "1m",
      "5m",
      "15m",
      "30m",
      "1h",
      "2h",
      "1d"
    ],
    "time_options": [
      "5m",
      "15m",
      "1h",
      "6h",
      "12h",
      "24h",
      "2d",
      "7d",
      "30d"
    ]
  },
  "timezone": "",
  "title": "54.236.37.243",
  "uid": "Bkj0YEJmz",
  "version": 2
}
```

[Load] - [Import]

Tips

You can stop grafana by command `docker stop grafana_monitor`.
You can start grafana by command `docker start grafana_monitor`.

Logs

```shell
docker logs -f influxdb_monitor
docker logs -f tron-monitor
docker logs -f grafana_monitor
```

## Show

![1.png](images/1.png)
![2.png](images/2.png)
![3.png](images/3.png)