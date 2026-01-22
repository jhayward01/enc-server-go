#!/bin/bash

# Integration Test #1
feclient --v2 | cut -f3- -d' ' | grep -v ^key > /tmp/test1.actual
docker logs enc-server-go-fe
docker logs enc-server-go-be
docker logs enc-server-go-mongodb-1
# diff /tmp/test1.actual test/test1.expect
