#!/bin/bash

# Integration Test #1
feclient --v2 | cut -f3- -d' ' | grep -v ^key > /tmp/test1.actual
sleep 15
docker logs enc-server-go-fe
docker logs enc-server-go-be
diff /tmp/test1.actual test/test1.expect
