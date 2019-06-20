import json
import sys

import requests
from dappctrl_rpc import *

offering_file_name = sys.argv[1]

ip_finder_url = "http://icanhazip.com"
country_finder_url = "http://ipinfo.io"

token = get_token(default_password)
print("\tToken: {}".format(token))

ip = requests.get(ip_finder_url).content
print("\nIP: {}".format(ip))

ip_info = requests.get("{}/{}".format(country_finder_url, ip))
country = ip_info.json()["country"]
print("Country: {}".format(country))

with open(offering_file_name) as f:
    offering = json.load(f)
offering["country"] = country

print("\nOffering: {}".format(json.dumps(offering, indent=8)))
