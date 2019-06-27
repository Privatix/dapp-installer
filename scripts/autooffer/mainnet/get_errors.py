from dappctrl_rpc import *

token = get_token(default_password)
print("\tToken: {}".format(token))

errors = get_logs(token, ["error"], "", "2017-1-1T1:1:1", "2217-1-1T1:1:1", 0, 100)
print("Errors: {}", errors)
