#!/bin/bash
KRBUSER=$(cat /secret/keytabvol/user)
USER=$(echo "$KRBUSER" | awk -F '@' '{print $1}')
export USER
kinit -kt /secret/keytabvol/keytab "${KRBUSER}"
python query_htc.py "$@"