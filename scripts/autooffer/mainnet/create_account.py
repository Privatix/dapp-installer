import os

from dappctrl_rpc import *

set_password(default_password)

token = get_token(default_password)
print("\tToken: {}".format(token))

account = create_account(token)
print("\tAccount: {}".format(account))

eth_address = get_eth_address(token, account)
print("\tEth address: 0x{}".format(eth_address))

private_key = export_private_key(token, account)
print("\tPrivate key: {}".format(private_key))

file_name = "{}.json".format(eth_address)
with open(file_name, 'w') as f:
    f.write(private_key)

print("\tPrivate key file: {}/{}".format(
    os.path.dirname(os.path.realpath(__file__)),
    file_name))
