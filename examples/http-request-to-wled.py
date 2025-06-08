#!/usr/bin/env python

import json

import requests


def wled_send_request(wled_addr: str, preset_id: int) -> int:
    """
    curl -X POST "http://${WLED_IP}/json/state" -d '{"on":"t"}' -H "Content-Type: application/json"
    """

    url = f"http://{wled_addr}/json/state"

    if preset_id == 0:
        data = json.dumps({"on": "t"})
    else:
        data = json.dumps({"ps": preset_id})

    headers = {"Content-Type": "application/json"}
    resp = requests.post(url, data, headers=headers)
    return resp.status_code


resp = wled_send_request("192.168.1.48", 0)
print(resp)
