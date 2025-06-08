#!/usr/bin/env bash

WLED_IP="192.168.1.48"

curl http://${WLED_IP}/json/state -d '{"on":"t","v":true}' -H "Content-Type: application/json"
