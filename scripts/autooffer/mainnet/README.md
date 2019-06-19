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

Ensure, that PRIX has been transferred to PSC:

```bash
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
	
Get public address
	Ok: <Response [200]>
	Eth address: 0x9486205adc7147ae551804c97c5bbb723ec7b826
	
Export private key
	Ok: <Response [200]>
	Private key: eyJhZGRyZXNzIjoiOTQ4NjIwNWFkYzcxNDdhZTU1MTgwNGM5N2M1YmJiNzIzZWM3YjgyNiIsImNyeXB0byI6eyJjaXBoZXIiOiJhZXMtMTI4LWN0ciIsImNpcGhlcnRleHQiOiI5MmUxZGVjZjVjMWVkNjg5ZWYyY2MyZGEzZWMzNGYzYmVmMGY3ZTdlNDk2MDkyNjVlNWNmZTEwNzBhZmE4NmU3IiwiY2lwaGVycGFyYW1zIjp7Iml2IjoiNDQyMGQwZDgyMzIyNThiZDQ0MzZjZDU3ZGM0ZGZjZWEifSwia2RmIjoic2NyeXB0Iiwia2RmcGFyYW1zIjp7ImRrbGVuIjozMiwibiI6MjYyMTQ0LCJwIjoxLCJyIjo4LCJzYWx0IjoiNzhiNzZlODA0MDA2M2M5NzJlYjIxMjFlNTZkMDBkZjJlYjhiNTM4YzkyNTlkZWU3ZGJjOWVlMTA3ZDQ5NzBkMCJ9LCJtYWMiOiJjYjJjM2Y5ZTUzODgwY2NiODQ4ZGZmMWE2YWE1M2JiYWY1Zjc5ZGJjYjZjMTQyMDVhNWFjMDhiNmJmMWZlYjczIn0sImlkIjoiMDE3MzNlNDEtMmUxYy00ZGM5LThhMmUtNTBjZjc4NDg1OTIyIiwidmVyc2lvbiI6M30=
	Private key file: /Users/andrei/tmp/9486205adc7147ae551804c97c5bbb723ec7b826.json
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
