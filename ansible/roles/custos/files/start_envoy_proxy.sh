#!/bin/bash
if ! docker ps --format "apachecustos/custos-rest-proxy:latest" | grep -w apachecustos/custos-rest-proxy &> /dev/null; then
  docker run -it -d --restart always  -v $(pwd)/logs:/var/log/envoy -e ENVOY_UID=777  -p 10000:50000    apachecustos/custos-rest-proxy:latest   -c /etc/envoy/envoy.yaml
fi