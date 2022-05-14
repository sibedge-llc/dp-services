# Docker

Eventer service requires to have kafka and postgres instances be run and configured at start

To run the service locally simply run supplied docker-compose

```shell
docker-compose up -d
```

This loads all reqiored instances and mount the service to 9099 port making accessible it by
http://localhost:9099 entry point see [USAGE](./USAGE.md) for details

