#!/bin/bash

# Integration Test #1
# feclient | cut -f3- -d' ' | grep -v ^key > /tmp/test1.actual
feclient | cut -f3- -d' ' > /tmp/test1.actual
diff /tmp/test1.actual test/test1.expect
