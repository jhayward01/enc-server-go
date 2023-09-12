#!/bin/bash

# Integration Test #1
make install-client-fe
feclient | cut -f3- -d' ' | grep -v ^key > /tmp/test1.actual
diff /tmp/test1.actual itest/test1.expect
