#!/bin/bash

# Integration Test #1
feclient | cut -f3- -d' ' | grep -v ^key > /tmp/fe-be-db.actual; \
diff /tmp/fe-be-db.actual test/fe-be-db.expect
