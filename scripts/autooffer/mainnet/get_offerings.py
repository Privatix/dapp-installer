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

    for offering in offerings:
        print("\n\t{} ({}):".format(offering["serviceName"], offering["id"]))
        print("\t\tstatus: {}".format(offering["status"]))
        print("\t\thash: {}".format(offering["hash"]))
        print("\t\tsupply: {}".format(offering["supply"]))
        print("\t\tcurrentSupply: {}".format(offering["currentSupply"]))
