# How to publish an offering

## Prepare

### Before build

- build for unix
- cd bin/installbuilder/linux-dapp-installer
- edit config "dapp-installer.config.json" from client to agent  

 
### Download

- app.tar.xz
- dapp-installer
- dapp-installer.config.json
- autooffer.py

## Install

Execute: 

```bash
sudo ./dapp-installer install --config dapp-installer.config.json
sudo python autooffer.py
```

## Notes

```
  log: /var/log/autooffer.log
 
  paths:
      - hardcoded
      - /var/lib/container/agent/product/73e17130-2a1d-4f7d-97a8-93a9aaa6f10d/config/adapter.config.json
      - /var/lib/container/agent/dappctrl/dappctrl.config.json
```

### Test

```bash
sudo ./dapp-installer remove --workdir /var/lib/container/agent/
sudo service network-manager restart
sudo ./dapp-installer install --config dapp-installer.config.json
sudo python autooffer.py
```


