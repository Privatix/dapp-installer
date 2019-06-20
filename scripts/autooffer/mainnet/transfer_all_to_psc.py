import time
from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

accounts = get_accounts(token)
for account in accounts:
    print("\nProcessing account: {} ({})".format(account["name"], account["id"]))
    transfer_tokens(token, account["id"], account["ptcBalance"], "psc")
    time.sleep(3)
    transactions = get_eth_transactions(token, "accountAggregated", account["id"], 0, 100)
    for transaction in transactions:
        print("\n\t{}:\n\t\t{} {}\n\t\thttps://etherscan.io/tx/0x{}".format(
            transaction["method"],
            transaction["status"],
            transaction["issued"],
            transaction["hash"]))
