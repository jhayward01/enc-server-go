#!/bin/bash

# Integration Test #1
feclient --v2 | cut -f3- -d' ' | grep -v configs > /tmp/test1.actual
diff /tmp/test1.actual test/test1.expect
