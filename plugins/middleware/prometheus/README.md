# 备注

直接使用 github.com/grpc-ecosystem/go-grpc-prometheus 实现


## docker-compose.yml
```
version: '2'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - 9090:9090
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  grafana:
    image: grafana/grafana
    ports:
      - 3000:3000
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=password
    volumes:
      - $PWD/extra/grafana_db:/var/lib/grafana grafana/grafana

```