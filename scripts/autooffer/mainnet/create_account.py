import os

import requests

# install_dependencies.sh

header = {
    'Content-Type': 'application/json',
}

default_password = "Qwerty=999"
endpoint = "http://localhost:8888/http"

private_key_file_name = "private_key.json"


def _check_ok(text, response):
    print(text)
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
    _check_ok("Get public address", response)

    return response.json()["result"]


set_password(default_password)

token = get_token(default_password)
print("\tToken: {}".format(token))

account = create_account(token)
print("\tAccount: {}".format(account))

eth_address = get_eth_address(token, account)
print("\tEth address: 0x{}".format(eth_address))

private_key = export_private_key(token, account)
print("\tPrivate key: {}".format(private_key))

with open(private_key_file_name, 'w') as f:
    f.write(private_key)

print("\tPrivate key file: {}/{}".format(os.path.dirname(os.path.realpath(__file__)), private_key_file_name))

