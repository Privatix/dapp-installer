import json

from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

products = get_products(token)
for product in products:
    offerings = get_agent_offerings(token, product["id"],
                                    ["empty", "registering", "registered",
                                     "popping_up", "popped_up", "removing",
                                     "removed"]
                                    , 0, 100)
    print("\tOfferings: {}".format(json.dumps(offerings, indent=8)))
