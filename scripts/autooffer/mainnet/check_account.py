import json

from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

accounts = get_accounts(token)
print("\tAccounts: {}".format(json.dumps(accounts, indent=8)))

for account in accounts:
    print("\nAccount: {} ({})".format(account["name"], account["id"]))
    transactions = get_eth_transactions(token, "accountAggregated", account["id"], 0, 100)
    for transaction in transactions:
        print("\n\t{}:\n\t\t{} {}\n\t\thttps://etherscan.io/tx/0x{}".format(
            transaction["method"],
            transaction["status"],
            transaction["issued"],
            transaction["hash"]))
