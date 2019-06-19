import json

from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

account = get_account(token)
print("\tAccount: {}".format(json.dumps(account, indent=8)))
