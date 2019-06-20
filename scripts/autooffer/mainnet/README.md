# mainnet tools for offering publication

## Offering's publication steps

### 0. Install prerequisites

```bash
./install_dependencies.sh
```

### 1. Create an account

```bash
python create_account.py
```

### 2. Transfer to them eth and PRIX

By using exchange or own wallet.

### 3. Check that founds has been delivered

```bash
python check_account.py
```

### 4. Transfer all PRIX from PTS to PSC 

```bash
python transfer_all_to_psc.py
```

Ensure, that PRIX has been transferred to PSC (usually it takes 5-10 min):

```bash
python get_transactions.py
python check_account.py
```

### 5. Publish an offering

```bash
python publish_offering.py ./offering.json
```

### 6. Transfer all PRIX from PSC to PTS  

```bash
python transfer_all_to_ptc.py
```

Ensure, that PRIX has been transferred to PTC (usually it takes 5-10 min):

```bash
python get_transactions.py
python check_account.py
```

## Tools

### create_account.py

#### Usage

```bash
python create_account.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	token: Bep9ISGVH6JD1KIQcyRZdRsCp_lf4e9BEv21iBqIMiI=
	
Generate account
	Ok: <Response [200]>
	Account: eec83276-bc94-4dc4-b04f-cc5e5173a6fb
	
Get eth address
	Ok: <Response [200]>
	Eth address: 0x9486205adc7147ae551804c97c5bbb723ec7b826
	
Export private key
	Ok: <Response [200]>
	Private key: Private key: {"address":"9486205adc7147ae551804c97c5bbb723ec7b826","crypto":{"cipher":"aes-128-ctr","ciphertext":"92e1decf5c1ed689ef2cc2da3ec34f3bef0f7e7e49609265e5cfe1070afa86e7","cipherparams":{"iv":"4420d0d8232258bd4436cd57dc4dfcea"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"78b76e8040063c972eb2121e56d00df2eb8b538c9259dee7dbc9ee107d4970d0"},"mac":"cb2c3f9e53880ccb848dff1a6aa53bbaf5f79dbcb6c14205a5ac08b6bf1feb73"},"id":"01733e41-2e1c-4dc9-8a2e-50cf78485922","version":3}
	Private key file: /Users/user/tmp/private_key.json
```

### check_account.py

#### Usage

```bash
python check_account.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	Token: cjx0AdPBH36habnp1N6lWJnb1hw_ixxau0Ydq0H-AeE=

Get account
	Ok: <Response [200]>
	Account: [
        {
                "ethAddr": "9486205adc7147ae551804c97c5bbb723ec7b826", 
                "name": "main", 
                "inUse": true, 
                "ptcBalance": 1000000000, 
                "lastBalanceCheck": "2019-06-19T17:11:57.227578+03:00", 
                "ethBalance": 50000000000000000, 
                "pscBalance": 0, 
                "id": "eec83276-bc94-4dc4-b04f-cc5e5173a6fb", 
                "isDefault": true
        }
]
```

### transfer_all_to_psc.py

#### Usage

```bash
python transfer_all_to_psc.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	Token: CGWMRsSDVJQ0Immnt90VlCBswDv0L6FGymRK1iBPgl4=

Get accounts
	Ok: <Response [200]>

Processing account: main (eec83276-bc94-4dc4-b04f-cc5e5173a6fb)

Transfer tokens (amount: 1000000000, gas price: 6000000000, direction: psc)
	Ok: <Response [200]>
```

### get_transactions.py

#### Usage

```bash
python get_transactions.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	Token: U0me_ixhZahaSCK2brwzQYfi7xz2OIseB-_j-zash1Y=

Get accounts
	Ok: <Response [200]>

Get eth transactions (type: accountAggregated, id: eec83276-bc94-4dc4-b04f-cc5e5173a6fb, offset: 0, limit: 100)
	Ok: <Response [200]>
--------------------------------------------------------------------------------

PSCAddBalanceERC20:
	sent 2019-06-20T12:10:30.705074+03:00
	https://etherscan.io/tx/0x91ad110fbb3ff0f2e32b7150d36ca6b1c9e8198b9da2561decb6c71933d3435c

PTCIncreaseApproval:
	sent 2019-06-20T12:10:01.500787+03:00
	https://etherscan.io/tx/0xc5bb8da80d7f68e1637c6d455e11a8c3cf6316d095dfd30fe09e1912d1c34a3e
```


### get_offerings.py

#### Usage

```bash
python get_offerings.py
```

#### Output

```
Get token
	Ok: <Response [200]>
	Token: iB60moN_beJFSRGIEY77NicS-gr31fHKsfpWLqpCb2s=

Get products
	Ok: <Response [200]>

Get agent offerings (product_id: 89e338bf-f594-4c6d-89fc-6ccda002cf26, status: ['empty', 'registering', 'registered', 'popping_up', 'popped_up', 'removing', 'removed'], offset: 0, limit: 100)
	Ok: <Response [200]>
	Offerings: {
        "totalItems": 1, 
        "items": [
                {
                        "somcSuccessPing": null, 
                        "somcType": 0, 
                        "supply": 30, 
                        "billingType": "postpaid", 
                        "agent": "9486205adc7147ae551804c97c5bbb723ec7b826", 
                        "billingInterval": 1, 
                        "unitType": "units", 
                        "somcData": "", 
                        "id": "11400304-a20c-4348-aa88-999d2d309631", 
                        "autoPopUp": true, 
                        "isLocal": false, 
                        "rawMsg": "eyJhZ2VudFB1YmxpY0tleSI6IkJOM0ZRU2VLS2pRQXhkVUM3dkdDZDJRd1VsYUZvQ29mUnFMU0RMSEJzVkRuTlF5RzlzUzVoa0dqYlBiUkRzSGZPZ3JFM1FCU2ZUTjVkUUt0cEMzXzQ3dz0iLCJ0ZW1wbGF0ZUhhc2giOiI5ZWE5NjU5YTk2OWRjYjkzMWM3ZjgxOWRlODNlNmUxNGRhNTMxNmY1YWJlM2M5NWFiOGM2MzhkNWZmOGRkNTFjIiwiY291bnRyeSI6IlJVIiwic2VydmljZVN1cHBseSI6MzAsInVuaXROYW1lIjoiTUIiLCJ1bml0VHlwZSI6InVuaXRzIiwiYmlsbGluZ1R5cGUiOiJwb3N0cGFpZCIsInNldHVwUHJpY2UiOjAsInVuaXRQcmljZSI6MTAwMCwibWluVW5pdHMiOjEwMDAwLCJtYXhVbml0IjozMDAwMCwiYmlsbGluZ0ludGVydmFsIjoxLCJtYXhCaWxsaW5nVW5pdExhZyI6MTAwLCJtYXhTdXNwZW5kVGltZSI6MTgwMCwibWF4SW5hY3RpdmVUaW1lU2VjIjoxODAwLCJmcmVlVW5pdHMiOjAsIm5vbmNlIjoiMTE0MDAzMDQtYTIwYy00MzQ4LWFhODgtOTk5ZDJkMzA5NjMxIiwic2VydmljZVNwZWNpZmljUGFyYW1ldGVycyI6ImV5SnRhVzVFYjNkdWJHOWhaRTFpYVhSeklqb2dNVEF3TENBaWJXbHVWWEJzYjJGa1RXSnBkSE1pT2lBNE1IMD0ifanU7oonyfly9wGllllnbP3BdTOBTSRyM9uSpE-cnob6L2gr6zV-PZy19MaEeg3Zl1u0WzD53cCmdEyHLVRf8BE=", 
                        "maxInactiveTimeSec": 1800, 
                        "maxSuspendTime": 1800, 
                        "serviceName": "VPN", 
                        "template": "ab6964b8-5586-4bed-a546-795e944af586", 
                        "blockNumberUpdated": 1, 
                        "setupPrice": 0, 
                        "status": "empty", 
                        "product": "89e338bf-f594-4c6d-89fc-6ccda002cf26", 
                        "hash": "1efb5586c5a6506047e555cfe0126fb076a7a7b3ae8d65fff0c71772dc04a98c", 
                        "description": "VPN", 
                        "unitName": "MB", 
                        "maxBillingUnitLag": 100, 
                        "currentSupply": 30, 
                        "unitPrice": 1000, 
                        "country": "RU", 
                        "freeUnits": 0, 
                        "minUnits": 10000, 
                        "additionalParams": {
                                "minDownloadMbits": 100, 
                                "minUploadMbits": 80
                        }, 
                        "maxUnit": 30000
                }
        ]
}```
