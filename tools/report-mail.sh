#!/bin/bash

LOG_FILE=$1
TO=$2

tail -n 200 ${LOG_FILE} | mail -s "Sunfish4 GA Report" ${TO}
