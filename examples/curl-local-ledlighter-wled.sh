#!/usr/bin/env bash

for ACTION in on off toggle toggle; do
    echo -e "Action: $ACTION \n"
    curl -v http://localhost:3080/api/v1/led/0/state/$ACTION
    echo -e "\n############################################\n"
    sleep 1
done
