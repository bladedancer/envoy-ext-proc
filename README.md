# External Processing Demo

Simple Envoy ext_proc to test loadbalancing across multipe ext_proc services.

## Build

    make

## Run

1) The target is webhook.site, get a GUID from there first.

2) In separate terminals launch two instances of the ext-proc service

    ./bin/extprocdemo --port 10001

    ./bin/extprocdemo --port 10002


2) Use func-e to launch Envoy (https://func-e.io/)

    func-e run -c config/envoy.yaml

3) Envoy is listening on port 8080, call that with the webhook.site GUID, e.g.

    curl http://localhost:8080/03e9d944-0431-4027-b969-7022fc34e576

