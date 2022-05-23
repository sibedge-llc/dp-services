# Docker

Eventer service requires to have kafka and postgres instances be run and configured at start

To run the service locally simply run supplied docker-compose

```shell
docker-compose up -d
```

This loads all required instances and mount the service to 9099 port making accessible it by
http://localhost:9099 entry point.

See [USAGE](./USAGE.md) for help using the eventer service

# K8s

To have it run in K8s it has to be followed by certain steps

1. Clone helm charts repository [github.com/sibedge-llc/dp-charts](https://github.com/sibedge-llc/dp-charts)
   See [github.com/sibedge-llc/dp-charts/README.md](https://github.com/sibedge-llc/dp-charts/README.md) for the next steps
```shell
git clone git@github.com:sibedge-llc/dp-charts.git & cd dp-charts/eventer
```

2. Check service available
```shell
curl http://localhost:9099/
{"result":"OK"}
```

3. See [USAGE](./USAGE.md) for help using the eventer service
