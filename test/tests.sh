#!/bin/bash

# Integration Test #1
feclient --v2 | cut -c 44- > /tmp/test1.actual 
diff /tmp/test1.actual test/test1.expect
