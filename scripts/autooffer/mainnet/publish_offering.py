import json
import sys

from dappctrl_rpc import *

offering_file_name = sys.argv[1]

token = get_token(default_password)
print("\tToken: {}".format(token))

products = get_products(token)
product = products[0]

accounts = get_accounts(token)
account = accounts[0]

print("\nUsed product: {}".format(product["name"]))
print("\nUsed account: {}".format(account["name"]))

with open(offering_file_name) as f:
    offering = json.load(f)

offering["country"] = product["country"]
offering["product"] = product["id"]
offering["template"] = product["offerTplID"]
offering["agent"] = account["id"]
offering["serviceName"] = product["name"]
offering["description"] = product["name"]

print("\nOffering: {}".format(json.dumps(offering, indent=8)))

offering_id = create_offering(token, offering)
print("\nOffering id: {}".format(offering_id))

# 11400304-a20c-4348-aa88-999d2d309631
