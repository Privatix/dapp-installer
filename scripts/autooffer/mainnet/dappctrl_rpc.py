import os

import requests

# rpc doc:
#   https://github.com/Privatix/dappctrl/blob/master/doc/ui/rpc.md

header = {
    'Content-Type': 'application/json',
}

default_password = os.environ.get("DAPP_PASSWORD", "Qwerty=999")
default_id = 1

endpoint = "http://localhost:8888/http"


def _check_ok(text, response):
    print("\n{}".format(text))

    if not response.ok:
        print("\tError: {0}({1})".format(response.text, response.reason))
        exit(1)

    response_in_json = response.json()

    if "error" in response_in_json:
        print("\tError: {0}".format(response_in_json["error"]))
        exit(1)

    print('\tOk: ' + str(response))


def _request_payload(method, args):
    return {
        'method': method,
        'params': args,
        'id': default_id,
    }


def gwei(raw_eth):
    return raw_eth * 1000000000


def prix(raw_prix):
    return float(raw_prix) / 10000000


def eth(raw_eth):
    return float(raw_eth) / 1000000000000000000


def set_password(password):
    data = _request_payload("ui_setPassword", [password])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Set password", response)


def get_token(password):
    data = _request_payload("ui_getToken", [password])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get token", response)

    return response.json()["result"]


def get_accounts(token):
    data = _request_payload("ui_getAccounts", [token])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get accounts", response)

    return response.json()["result"]


def create_account(token, name="main"):
    data = _request_payload("ui_generateAccount", [token,
                                                   {
                                                       'name': name,
                                                       'isDefault': True,
                                                       'inUse': True,
                                                   }])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Generate account (name: {})".format(name), response)

    return response.json()["result"]


def get_object(token, object_type, object_id):
    data = _request_payload("ui_getObject", [token, object_type, object_id])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get object (type: {}, id: {})".format(object_type, object_id), response)

    return response.json()["result"]


def export_private_key(token, account_id):
    data = _request_payload("ui_exportPrivateKey", [token, account_id])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Export private key (account: {})".format(account_id), response)

    return response.json()["result"]


def transfer_tokens(token, account_id, token_amount, direction, gas_price=gwei(6)):
    data = _request_payload("ui_transferTokens", [token, account_id, direction,
                                                  token_amount, gas_price])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Transfer tokens (amount: {}, gas price: {}, direction: {})".format(token_amount, gas_price, direction),
              response)


# type: offering, channel, endpoint, account, accountAggregated
def get_eth_transactions(token, type, related_id, offset, limit):
    data = _request_payload("ui_getEthTransactions", [token, type, related_id,
                                                      offset, limit])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get eth transactions (type: {}, id: {}, offset: {}, limit: {})".format(type, related_id, offset, limit),
              response)

    return response.json()["result"]["items"]


def create_offering(token, offering):
    data = _request_payload("ui_createOffering", [token, offering])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Create offering", response)

    return response.json()["result"]


def get_products(token):
    data = _request_payload("ui_getProducts", [token])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get products", response)

    return response.json()["result"]


def get_agent_offerings(token, product_id, status, offset, limit):
    data = _request_payload("ui_getAgentOfferings", [token, product_id, status,
                                                     offset, limit])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok(
        "Get agent offerings (product_id: {}, status: {}, offset: {}, limit: {})".format(product_id, status, offset,
                                                                                         limit), response)
    return response.json()["result"]["items"]


def change_offering_status(token, offering_id, action, gas_price=gwei(6)):
    data = _request_payload("ui_changeOfferingStatus", [token, offering_id,
                                                        action, gas_price])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok(
        "Change offering status (offering_id: {}, action: {})".format(offering_id, action), response)


def get_logs(token, levels, text, lower_bound, upper_bound, offset, limit):
    data = _request_payload("ui_getLogs", [token, levels, text, lower_bound, upper_bound, offset, limit])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok(
        "Get logs (levels: {}, text: \"{}\", lower_bound: {}, upper_bound: {}, offset: {}, limit: {})".format(levels,
                                                                                                              text,
                                                                                                              lower_bound,
                                                                                                              upper_bound,
                                                                                                              offset,
                                                                                                              limit),
        response)
    return response.json()["result"]["items"]


def update_balance(token, account_id):
    data = _request_payload("ui_updateBalance", [token, account_id])

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Update balance (account_id: {})".format(account_id), response)
