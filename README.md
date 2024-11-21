# External Processing Demo

Simple Envoy ext_proc to test failure processing over multipe ext_proc services.

## Build Local

    make local.build

## Run Local

1) In separate terminals launch two instances of the ext-proc service

    ./bin/extprocdemo --port 10001 --successPercentage 100

    ./bin/extprocdemo --port 10002 --successPercentage 50


2) Use func-e to launch Envoy (https://func-e.io/)

    func-e run -c config/envoy.yaml

3) Envoy is listening on port 8080, call that with the webhook.site GUID, e.g.

    curl http://localhost:8080/03e9d944-0431-4027-b969-7022fc34e576


## Build Container

Change to your own dockerhub repo prefix. Alternatively don't build it and just pull mine.

    make docker.push

## Run In Cluster

1) Install the services

    kubectl create namespace envoy

    helm upgrade -i -n envoy extproc ./helm/extproc

    helm upgrade -i -n envoy envoy ./helm/envoy

2) Port forward

    kubectl port-forward -n envoy --address 0.0.0.0 svc/envoy 8080:8080

3) Call the service

    curl http://localhost:8080/03e9d944-0431-4027-b969-7022fc34e576