#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
        Automatic offering publish
        on pure Python 2.7

"""

import sys
import logging
from codecs import open
from time import time, sleep
from threading import Thread
from urllib2 import urlopen, Request
from os.path import isfile
from json import dump, dumps, load, loads

main_conf = dict(
    log_path='/var/log/autooffer.log',
)

logging.getLogger().setLevel('DEBUG')
form_console = logging.Formatter(
    '%(message)s',
    datefmt='%m/%d %H:%M:%S')

form_file = logging.Formatter(
    '%(levelname)7s [%(lineno)3s] %(message)s',
    datefmt='%m/%d %H:%M:%S')

fh = logging.FileHandler(main_conf['log_path'])  # file debug
fh.setLevel('DEBUG')
fh.setFormatter(form_file)
logging.getLogger().addHandler(fh)


ch = logging.StreamHandler()  # console debug
ch.setLevel('INFO')
ch.setFormatter(form_console)
logging.getLogger().addHandler(ch)
logging.debug('\n\n\n--- Begin ---')


class Init:
    waiting = True
    in_args = None  # arguments with which the script was launched.

    @staticmethod
    def long_waiting():
        symb = ['|', '/', '-', '\\']

        while Init.waiting:
            for i in symb:
                sys.stdout.write("\r[%s]" % (i))
                sys.stdout.flush()
                if not Init.waiting:
                    break
                sleep(0.3)
                if not Init.waiting:
                    break

        sys.stdout.write("\r")
        sys.stdout.write("")
        sys.stdout.flush()
        Init.waiting = True

    @staticmethod
    def wait_decor(func):
        def wrap(obj, args=None):
            logging.debug('Wait decor args: {}.'.format(args))
            st = Thread(target=Init.long_waiting)
            st.daemon = True
            st.start()
            if args:
                res = func(obj, args)
            else:
                res = func(obj)
            Init.waiting = False
            sleep(0.5)
            return res

        return wrap

    @staticmethod
    def file_rw(p, w=False, data=None, log=None, json_r=False):
        try:
            if log:
                logging.debug('{}. Path: {}'.format(log, p))

            if w:
                f = open(p, 'w')
                if data:
                    if json_r:
                        dump(data, f, indent=4)
                    else:
                        f.writelines(data)
                f.close()
                return True
            else:
                f = open(p, 'r')
                if json_r:
                    if f:
                        data = load(f)
                else:
                    data = f.readlines()
                f.close()
                return data
        except BaseException as rwexpt:
            logging.error('R/W File: {}'.format(rwexpt))
            return False

    @staticmethod
    def conf_dappctrl_json(AutoOffer):

        p = '/var/lib/container/agent/dappctrl/dappctrl.config.json'

        # Read dappctrl.config.local.json
        data = Init.file_rw(p=p, json_r=True, log='Read dappctrl conf')
        if not data:
            logging.error('Failed to read dappctrl.config.json')

        data['StaticPassword'] = AutoOffer.pswd

        # Rewrite dappctrl.config.local.json
        Init.file_rw(p=p, w=True, json_r=True, data=data,
                     log='Rewrite dappctrl.config.json')


class AutoOffer():
    def __init__(self):
        self.id = 1
        self.url = 'http://localhost:8888/http'
        self.pswdSymbol = 12
        self.acc_name = 'TestAcc'
        self.botUrl = 'http://89.38.96.53:3000/getprix'
        self.botAuth = 'dXNlcjpoRmZWRWRVMkNva0Y='
        self.offerData = None
        self.pswd = 'Qwerty=999'
        self.token = None
        self.ethAddr = None
        self.prixHash = None
        self.agent_id = None  # id of account to be created.
        # Session.Product field from adapter.config.json. Specifies for which product offering should be published.
        self.product_id = None
        self.ethHash = None
        self.ptcBalance = None
        self.pscBalance = None
        self.ethBalance = None
        self.offer_id = None
        self.gasPrice = 2000000000
        self.waitBot = 1
        self.waitblockchain = 90
        self.vpnConf = '/var/lib/container/agent/product/73e17130-2a1d-4f7d-97a8-93a9aaa6f10d/config/adapter.config.json'
        self.template_id=""

    def _getAgentOffer(self, mark):
        logging.info('Get Offerings. Mark: {}'.format(mark))
        # Get Offerings For Agent
        data = {
            'method': 'ui_getAgentOfferings',
            'params': [
                self.token,
                self.product_id,
                ['registered'],
                0,
                1
            ],
            'id': self.id,
        }
        timeWait = 25 * 60
        timeStar = time()
        while time() - timeStar < timeWait:
            res = self.__urlOpen(data=data, key='result')
            if res[0]:
                items = res[1].get('items')
                logging.debug("items: {}".format(items))
                if items and isinstance(items, (list, set, tuple)):
                    status = items[0].get('status')
                    logging.debug('Offer status: {}'.format(status))
                    if status == 'registered':
                        logging.debug('Offerings for agent exist.')
                        return True, 'All done'
            logging.debug('Wait')
            sleep(60)
        logging.info('Does not exist offerings for agent.')
        if not mark:
            logging.info('Try again.')
            return self._statusOffer(mark=True)
        return False, res[1]

    def __getProductId(self):
        logging.info('Get Product Id')
        if isfile(self.vpnConf):
            try:
                f = open(self.vpnConf)
                raw_data = loads(f.read())
                logging.debug('Read vpn conf: {}'.format(raw_data))
                self.product_id = raw_data['Sess']['Product']
                logging.debug('Product id: {}'.format(self.product_id))
                return True, 'Product id was found'

            except BaseException as readexpt:
                logging.error('Read vpn conf: {}'.format(readexpt))
                return False, readexpt

        return False, 'there is no {} to determine Product Id'.format(
            self.vpnConf)

    def __checkOfferData(self):
        logging.debug('Check offer data')
        params = {
            "product": self.product_id,
            "template": self.template_id,
            "agent": self.agent_id,
            "serviceName": "name",
            "description": "description",
            "country": self.__getCountryName(),
            "supply": 30,
            "unitName": "MB",
            "autoPopUp": True,
            "unitType": "units",
            "billingType": "postpaid",
            "setupPrice": 0,
            "unitPrice": 1000,
            "minUnits": 10000,
            "maxUnit": 30000,
            "billingInterval": 1,
            "maxBillingUnitLag": 100,
            "maxSuspendTime": 1800,
            "maxInactiveTimeSec": 1800,
            "freeUnits": 0,
            "additionalParams": {"minDownloadMbits": 100,
                                 "minUploadMbits": 80},
        }

        if self.offerData:
            logging.debug('From file: {}'.format(self.offerData))
            res = self.offerData.get('country')
            if res:
                if not params['country'].lower() == res.lower():
                    logging.info('You country name {} from config : {}\n'
                                 'does not match with country calculated '
                                 'by your IP : {}\n.'.format(
                                     res, self.in_args['file'], params['country']))
                    logging.info('Choose which country to use 1 or 2:\n'
                                 '1 - {}\n'
                                 '2 - {}'.format(res, params['country']))
                    choise_task = {1: res, 2: params['country']}

                    while True:
                        choise_code = raw_input('>')

                        if choise_code.isdigit() and int(
                                choise_code) in choise_task:
                            self.offerData['country'] = choise_task[
                                int(choise_code)]
                            break
                        else:
                            logging.info(
                                'Wrong choice. Make a choice between: '
                                '{}'.format(choise_task.keys()))

            params.update(self.offerData)

        else:
            logging.info(
                'You file offer is empty.Install with default params')

        return params

    def validateJson(self, path):
        logging.info('Checking JSON')
        try:
            with open(path) as f:
                try:
                    self.offerData = load(f)
                    logging.debug('Json is valid: {}'.format(self.offerData))
                    return True, 'Json is valid'
                except ValueError as e:
                    logging.error('Read file: {}'.format(e))
                    return False, 'This is not json format.' \
                                  'Perhaps you are using a single quote \', ' \
                                  'instead of a double quote ".Check your structure.'
        except BaseException as oexpt:
            logging.error('Open file: {}'.format(oexpt))
            return False, 'Trouble when try open file: {}. Error: {}'.format(
                path, oexpt)

    @Init.wait_decor
    def republishOffer(self):
        logging.debug('Republish')
        res = self.__getProductId()
        if res[0]:
            res = self._getAcc()
            logging.debug('Get Acc: {}'.format(res))
            if res[0]:

                self.ethAddr = res[1][0]['ethAddr']
                self.ptcBalance = res[1][0]['ptcBalance']
                self.pscBalance = res[1][0]['pscBalance']
                self.ethBalance = res[1][0]['ethBalance']
                self.agent_id = res[1][0]['id']

                logging.debug('ethAddr: {}'.format(self.ethAddr))
                logging.debug('agent_id: {}'.format(self.agent_id))
                logging.debug('ethBalance: {}'.format(self.ethBalance))
                logging.debug('pscBalance: {}'.format(self.pscBalance))
                logging.debug('ptcBalance: {}'.format(self.ptcBalance))
                res = self._askBot()
                if not res[0]:
                    return res

                self._wait_blockchain(target='ptc', republ=True)
                res = self._transfer()
                if not res[0]:
                    return res
                self._wait_blockchain(target='psc', republ=True)
                res = self._createOffer()

                logging.debug('ethAddr: {}'.format(self.ethAddr))
                logging.debug('agent_id: {}'.format(self.agent_id))
                logging.debug('ethBalance: {}'.format(self.ethBalance))
                logging.debug('pscBalance: {}'.format(self.pscBalance))
                logging.debug('ptcBalance: {}'.format(self.ptcBalance))
                logging.debug('product_id: {}'.format(self.product_id))
                logging.debug('ethHash: {}'.format(self.ethHash))
                logging.debug('offer_id: {}'.format(self.offer_id))
                logging.debug('prixHash: {}'.format(self.prixHash))
                logging.debug('gasPrice: {}'.format(self.gasPrice))
                if not res[0]:
                    return res
                return self._statusOffer()
        return res

    @Init.wait_decor
    def offerRun(self):
        res = self.__getProductId()
        if not res[0]:
            return res
        res = self._setPswd()
        if res[0]:
            logging.debug('Eth addr: {}'.format(self.ethAddr))
            logging.debug('Generate acc id: {}'.format(self.agent_id))

            res = self._askBot()
            if not res[0]:
                return res

            self._wait_blockchain(target='ptc')
            res = self._transfer()
            if not res[0]:
                return res
            self._wait_blockchain(target='psc')
            template = self._get_template()
            self.template_id = template['id']
            res = self._createOffer()
            logging.debug('product_id: {}'.format(self.product_id))
            logging.debug('offer_id: {}'.format(self.offer_id))
            logging.debug('ethBalance: {}'.format(self.ethBalance))
            logging.debug('pscBalance: {}'.format(self.pscBalance))
            logging.debug('ptcBalance: {}'.format(self.ptcBalance))
            logging.debug('ethHash: {}'.format(self.ethHash))
            logging.debug('agent_id: {}'.format(self.agent_id))
            logging.debug('prixHash: {}'.format(self.prixHash))
            logging.debug('ethAddr: {}'.format(self.ethAddr))
            logging.debug('gasPrice: {}'.format(self.gasPrice))
            if not res[0]:
                return res
            return self._statusOffer()

        else:
            return res

    def __getCountryName(self):
        country = None
        try:
            ip = urlopen('http://icanhazip.com').read()
            raw_data = urlopen(
                'http://ipinfo.io/{}'.format(ip)).read()
            country = loads(raw_data)['country']
        except BaseException as cntr:
            logging.debug('Error when try get cuntry: {}'.format(cntr))
            country = raw_input(
                prompt='Please enter your country name, abbreviated. For example US.')
        finally:
            return country

    def _statusOffer(self, mark=False):
        logging.info('Offering status')
        data = {
            'method': 'ui_changeOfferingStatus',
            'params': [
                self.token,
                self.offer_id,
                'publish',
                self.gasPrice,

            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data)
        if res[0]:
            return self._getAgentOffer(mark)
        else:
            return False, res[1]

    def _get_template(self):
        logging.info('Get template')
        data = {
            'method': 'ui_getTemplates',
            'params': [
                self.token,
                "offer"
            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        return res[1][0]


    def _createOffer(self):
        logging.info('Offering create')
        data = {
            'method': 'ui_createOffering',
            'params': [
                self.token,
                self.__checkOfferData()
            ],
            'id': self.id,
        }

        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.offer_id = res[1]
            return True, res[1]
        else:
            return False, res[1]

    def _wait_blockchain(self, target, republ=False):
        logging.info('Wait blockchain.Target is: {}.'.format(target))
        waitCounter = 0
        while True:
            sleep(self.waitblockchain)
            res = self._getEth()
            logging.debug('Wait {} min'.format(waitCounter))
            waitCounter += 1
            if res[0]:
                if target == 'ptc' and int(res[1].get('ptcBalance', '0')):
                    if republ and int(self.ptcBalance) >= int(
                            res[1]['ptcBalance']):
                        continue
                    self.ptcBalance = res[1]['ptcBalance']
                    self.ethBalance = res[1]['ethBalance']
                    break
                elif target == 'psc' and int(res[1].get('pscBalance', '0')):
                    if republ and int(self.pscBalance) >= int(
                            res[1]['pscBalance']):
                        continue
                    self.pscBalance = res[1]['pscBalance']
                    self.ptcBalance = res[1]['ptcBalance']
                    self.ethBalance = res[1]['ethBalance']
                    break

    def _transfer(self):
        # Transfer some PRIX from PTC balance to PSC balance
        logging.info('Transfer PRIX')
        data = {
            'method': 'ui_transferTokens',
            'params': [
                self.token,
                self.agent_id,
                'psc',
                self.ptcBalance,
                self.gasPrice
            ],
            'id': self.id,
        }

        res = self.__urlOpen(data)
        if res[0]:
            return True, 'Ok'
        else:
            return False, res[1]

    def __urlOpen(self, data, key=None, url=None, auth=None):
        try:
            url = self.url if not url else url
            logging.debug('Request: {}'.format(data))
            request = Request(url)
            request.add_header('Content-Type', 'application/json')
            if auth:
                request.add_header('Authorization', "Basic {}".format(auth))
            response = urlopen(request, dumps(data))
            response = response.read()
            logging.debug('Response: {0}'.format(response))
            try:
                response = loads(response)
            except BaseException as jsonExpt:
                logging.error(jsonExpt)
                return False, jsonExpt

            if response.get('error', False):
                logging.error(
                    "Error on response: {}".format(response['error']))

                return False, response['error']
            if key:
                logging.debug('Get key: {}'.format(key))
                response = response.get(key, False)
                if not response:
                    logging.error('Key {} not exist in response'.format(key))
                    return False, 'Key {} not exist in response'.format(key)

            logging.debug('Response OK: {}'.format(response))
            return True, response

        except BaseException as urlexpt:
            logging.error('Url Exept: {}'.format(urlexpt))
            return False, urlexpt

    def _setPswd(self):
        # 1.Set password for UI API access
        logging.info('Set password')

        data = {
            'method': 'ui_setPassword',
            'params': [self.pswd],
            'id': self.id,
        }

        res = self.__urlOpen(data)
        if res[0]:
            return self._getTok()
        else:
            return False, res[1]

    def _getTok(self):
        # Given paswd and returns new access token.
        logging.info('Get token')

        data = {
            'method': 'ui_getToken',
            'params': [self.pswd],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.token = res[1]
            logging.debug('Token: {}'.format(self.token))
            return self._createAcc()
        else:
            return False, res[1]

    def _getAcc(self):
        # Create account
        '''curl -X POST -H "Content-Type: application/json" --data '{"method": "ui_getAccounts", "params": ["qwert"], "id": 67}' http://localhost:8888/http

        {"jsonrpc":"2.0",
        "id":67,
        "result":[{"id":"25b84988-fde6-4882-91bc-ab1dfb86cbdb","ethAddr":"35218b6fc288e093d55295e2c3ce7304d216be64","isDefault":true,"inUse":true,"name":"TestAcc","ptcBalance":0,"pscBalance":700000000,"ethBalance":49654270000000000,"lastBalanceCheck":"2018-11-29T14:24:41.730652+01:00"}]}

        '''
        logging.info('Get account')
        data = {
            'method': 'ui_getAccounts',
            'params': [self.token],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        return res

    def _createAcc(self):
        # Create account
        logging.info('Create account')
        data = {
            'method': 'ui_generateAccount',
            'params': [
                self.token,
                {
                    'name': self.acc_name,
                    'isDefault': True,
                    'inUse': True,
                }
            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            if isinstance(res[1], unicode):
                self.agent_id = res[1].encode()
            else:
                self.agent_id = res[1]

            return self._getEth()
        else:
            return False, res[1]

    def _getEth(self):
        # Get ethereum address of newly created account
        logging.debug('Get ethereum address')
        data = {
            'method': 'ui_getObject',
            'params': [
                self.token,
                'account',
                self.agent_id,
            ],
            'id': self.id,
        }
        res = self.__urlOpen(data=data, key='result')
        if res[0]:
            self.ethAddr = res[1]['ethAddr']
            return res
        else:
            return False, res[1]

    def _askBot(self):
        # Ask Privatix bot to transfer PRIX and ETH to address of account in
        logging.info('Ask Privatix Bot')
        stop_mark = 0

        data = {
            'address': '0x{}'.format(self.ethAddr),
        }
        while True:
            res = self.__urlOpen(data=data, url=self.botUrl,
                                 auth=self.botAuth)
            if res[0]:
                if res[1].get('code') and res[1]['code'] == 200:
                    self.prixHash = res[1].get('prixHash')
                    self.ethHash = res[1].get('ethHash')
                    if self.prixHash and self.ethHash:
                        return True, 'OK'
                    logging.debug(
                        'prixHash:{}, ethHash:{}'.format(self.prixHash,
                                                         self.ethHash))
                else:
                    logging.error('Bot error: {}'.format(res[1]))
                stop_mark += 1
                sleep(self.waitBot)
            else:
                logging.error('Bot not answer: {}'.format(res[1]))
                return False, res[1]

            if stop_mark > 5:
                logging.error('Error when try ask to bot')
                return False, res[1]


if __name__ == '__main__':

    offer = AutoOffer()
    Init.conf_dappctrl_json(offer)
    res = offer.offerRun()
    if not res[0]:
        logging.error(
            'Auto offer: {}'.format(res[1]))
        raise BaseException(res[1])

    mess = '    Congratulations, you posted your offer!\n' \
        '    It will be published once an hour.\n' \
        '    Your ethereum address: 0x{}\n' \
        '    Your pasword : {}\n' \
        '    Please press enter to finalize the application.'.format(
            offer.ethAddr, offer.pswd)
    logging.info(mess)
