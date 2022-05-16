# Docker

Eventer service requires to have kafka and postgres instances be run and configured at start

To run the service locally simply run supplied docker-compose

```shell
docker-compose up -d
```

This loads all reqiored instances and mount the service to 9099 port making accessible it by
http://localhost:9099 entry point see [USAGE](./USAGE.md) for details

# K8s

To have it run in K8s it has to be followed by certain steps

1. Install and use `helm` cli tool according to you system
2. Use dedicated repository to download and access dedicated helm charts
```shell
git clone git@github.com:sibedge-llc/dp-charts.git
```
3. Install kafka/zookeeper to the k8s cluster
```shell
cd dp-charts/kafka
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install kafka bitnami/kafka \
  --namespace kafka \
  --create-namespace \
  -f values.yaml
```
4. Install postgresql to the k8s cluster
```shell
cd dp-charts/postgresql
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgresql bitnami/postgresql \
  --namespace postgresql \
  --create-namespace \
  -f values.yaml
```
5. Install eventer to the k8s cluster
```shell
cd dp-charts
helm upgrade --install --debug eventer eventer/ \
  --namespace eventer \
  --create-namespace \
  -f eventer/values.yaml
```

6. Mount service port to your host (in case you run the k8s locally)
```shell
kubectl port-forward service/eventer 9099:9099 -n eventer
```

7. Check service available
```shell
curl http://localhost:9099/
{"result":"OK"}
```
