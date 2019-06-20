import json

from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

accounts = get_accounts(token)
print("\tAccounts: {}".format(json.dumps(accounts, indent=8)))
