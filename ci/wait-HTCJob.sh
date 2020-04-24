#!/bin/bash

while test "$(kubectl get htcjob zero -o yaml|grep '^  succeeded:')" != "  succeeded: 1"
do
    sleep 10
done
