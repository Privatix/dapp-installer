import requests

header = {
    'Content-Type': 'application/json',
}

default_password = "Qwerty=999"
default_id = 1

endpoint = "http://localhost:8888/http"


def _check_ok(text, response):
    print("\n{}".format(text))
    if not response.ok:
        print("\tError: {0}({1})".format(response.text, response.reason))
        exit(1)
    print('\tOk: ' + str(response))


def set_password(password):
    data = {
        'method': "ui_setPassword",
        'params': [password],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Set password", response)


def get_token(password):
    data = {
        'method': "ui_getToken",
        'params': [password],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get token", response)

    return response.json()["result"]


def get_accounts(token):
    data = {
        'method': "ui_getAccounts",
        'params': [
            token
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get accounts", response)

    return response.json()["result"]


def create_account(token, name="main"):
    data = {
        'method': 'ui_generateAccount',
        'params': [
            token,
            {
                'name': name,
                'isDefault': True,
                'inUse': True,
            }
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Generate account (name: {})".format(name), response)
    return response.json()["result"]


def get_eth_address(token, account):
    data = {
        'method': 'ui_getObject',
        'params': [
            token,
            'account',
            account,
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get public address (account: {})".format(account), response)
    return response.json()["result"]["ethAddr"]


def export_private_key(token, account):
    data = {
        'method': 'ui_exportPrivateKey',
        'params': [
            token,
            account
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Export private key (account: {})".format(account), response)

    return response.json()["result"]


def transfer_tokens(token, account, token_amount, direction, gas_price=6000000000):
    data = {
        'method': 'ui_transferTokens',
        'params': [
            token,
            account,
            direction,
            token_amount,
            gas_price
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Transfer tokens (amount: {}, gas price: {}, direction: {})".format(token_amount, gas_price, direction),
              response)


def get_eth_transactions(token, type, related_id, offset, limit):
    data = {
        'method': 'ui_getEthTransactions',
        'params': [
            token,
            type,
            related_id,
            offset,
            limit,
        ],
        'id': default_id,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get eth transactions (type: {}, id: {}, offset: {}, limit: {})".format(type, related_id, offset, limit),
              response)

    return response.json()["result"]["items"]
