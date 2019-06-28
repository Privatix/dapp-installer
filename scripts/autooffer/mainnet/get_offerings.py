from dappctrl_rpc import *

token = get_token(default_password)

products = get_products(token)
for product in products:
    offerings = get_agent_offerings(token, product["id"],
                                    ["empty", "registering", "registered",
                                     "popping_up", "popped_up", "removing",
                                     "removed"]
                                    , 0, 100)

    for offering in offerings:
        print("-" * 80)
        print("\n{}:\n\tHash: 0x{}\n\tStatus: {}\n\n\tSupply: {}\n\tCurrent supply: {}\n\n\tId: {}".format(
            offering["serviceName"],
            offering["hash"],
            offering["status"],
            offering["supply"],
            offering["currentSupply"],
            offering["id"],
        ))
