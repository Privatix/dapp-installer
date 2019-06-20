# Tools for step-by-step offering publications

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

Ensure, that the offering has been published (usually it takes 5-10 min):

```bash
python get_offerings.py
```

### 6. Transfer all earned PRIX from PSC to PTS  

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

### check_accounts.py

#### Usage

```bash
python check_accounts.py
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
	Token: Rt-Xb27IqCZ-va-j8zQ6GzefEIWwrJR8EIyYu-aFbqY=

Get products
	Ok: <Response [200]>

Get agent offerings (product_id: 89e338bf-f594-4c6d-89fc-6ccda002cf26, status: ['empty', 'registering', 'registered', 'popping_up', 'popped_up', 'removing', 'removed'], offset: 0, limit: 100)
	Ok: <Response [200]>

	VPN (ec4a47f7-3c56-4cd3-b3c4-41553f3cf6f1):
		status: registered
		hash: d902ddbcdefa0c924bee6a41825c19b0cd00e3153c68abbd930e43e7b00401d8
		supply: 30
		currentSupply: 30

	VPN (11400304-a20c-4348-aa88-999d2d309631):
		status: empty
		hash: 1efb5586c5a6506047e555cfe0126fb076a7a7b3ae8d65fff0c71772dc04a98c
		supply: 30
		currentSupply: 30
```


### publish_offering.py

#### Usage

```bash
python publish_offering.py offering.json
```

#### Output

```
Get token
	Ok: <Response [200]>
	Token: 4EWTKhTbP1zhHnhzQl45i7ixLHI0FTM6I9Tse2Ksal0=

Get products
	Ok: <Response [200]>

Get accounts
	Ok: <Response [200]>

Used product: VPN

Used account: main

Offering: {
        "billingType": "postpaid", 
        "maxInactiveTimeSec": 1800, 
        "autoPopUp": true, 
        "description": "VPN", 
        "unitName": "MB", 
        "unitPrice": 1000, 
        "maxBillingUnitLag": 100, 
        "supply": 30, 
        "freeUnits": 0, 
        "agent": "eec83276-bc94-4dc4-b04f-cc5e5173a6fb", 
        "maxSuspendTime": 1800, 
        "product": "89e338bf-f594-4c6d-89fc-6ccda002cf26", 
        "billingInterval": 1, 
        "unitType": "units", 
        "serviceName": "VPN", 
        "template": "ab6964b8-5586-4bed-a546-795e944af586", 
        "minUnits": 10000, 
        "additionalParams": {
                "minDownloadMbits": 100, 
                "minUploadMbits": 80
        }, 
        "country": "RU", 
        "setupPrice": 0, 
        "maxUnit": 30000
}

Create offering
	Ok: <Response [200]>

Offering id: b706c092-dbcf-4847-b0f6-cf6c095d2cdd

Change offering status (offering_id: b706c092-dbcf-4847-b0f6-cf6c095d2cdd, action: publish)
	Ok: <Response [200]>
```
