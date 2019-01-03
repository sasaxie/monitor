# Monitor

## A Monitor for java-tron

Monitor is a open source monitor for java-tron. It's useful for monitoring nodes
 of java-tron.

## Features

- Node block number does not update alarm;
- Witness lost block alarm.
- Witness list change alarm.
- Proposed value change alarm.
- Witness lost block daily report.

## Getting Started

```shell
docker pull sasaxie/tron-monitor

docker run -d \
  --name docker-influxdb-grafana-monitor \
  -p 3003:3003 \
  -p 3004:8083 \
  -p 8086:8086 \
  -p 22022:22 \
  -p 8080:8080 \
  -v $PWD/influxdb:/var/lib/influxdb \
  -v $PWD/grafana:/var/lib/grafana \
  -v $PWD/monitor:/root/go/bin/conf \
  sasaxie/tron-monitor
```

## Grafana

Open <http://localhost:3003>

```
Username: root
Password: root
```

Modify /etc/grafana/grafana.ini root_url=http://[your ip]:3003

## InfluxDB

### Web Interface

Open <http://localhost:3004>

```
Username: root
Password: root
Port: 8086
```

## Configuration

nodes.json

```json
{
  "addresses": [
    {
      "ip": "127.0.0.1",
      "grpcPort": 50051,
      "httpPort": 8090,
      "type": "full_node",
      "tag": "MainNet",
      "monitor": "NowBlock,BlockMissed"
    }
  ]
}
```

Type:

- full_node
- mti_full_node
- witness_node
- sr_witness_node
- sr_witness_b_node
- gr_witness_node
- solidity_node

Monitor type:

- NowBlock
- BlockMissed

## Show

![node-detail.png](images/node-detail.png)