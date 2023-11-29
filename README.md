# s-tui-exporter

Little prom exporter written in go.

Connects to a host specified using env vars REMOTE_USERNAME, REMOTE_HOST and REMOTE_PASSWORD and scrapes s-tui json output for package power, serving it at its metrics endpoint (:8080/metrics) as prox_power prom metric.

