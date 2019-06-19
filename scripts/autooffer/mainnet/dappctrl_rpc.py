import requests

header = {
    'Content-Type': 'application/json',
}

default_password = "Qwerty=999"
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
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Set password", response)


def get_token(password):
    data = {
        'method': "ui_getToken",
        'params': [password],
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get token", response)

    return response.json()["result"]


def get_account(token):
    data = {
        'method': "ui_getAccounts",
        'params': [
            token
        ],
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get accounts", response)

    return response.json()["result"]


def create_account(token):
    data = {
        'method': 'ui_generateAccount',
        'params': [
            token,
            {
                'name': 'main',
                'isDefault': True,
                'inUse': True,
            }
        ],
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Generate account", response)
    return response.json()["result"]


def get_eth_address(token, account):
    data = {
        'method': 'ui_getObject',
        'params': [
            token,
            'account',
            account,
        ],
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Get public address", response)
    return response.json()["result"]["ethAddr"]


def export_private_key(token, account):
    data = {
        'method': 'ui_exportPrivateKey',
        'params': [
            token,
            account
        ],
        'id': 1,
    }

    response = requests.post(endpoint, json=data, headers=header)
    _check_ok("Export private key", response)

    return response.json()["result"]
