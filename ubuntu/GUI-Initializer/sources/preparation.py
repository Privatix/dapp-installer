#!/usr/bin/python

import requests
from pexpect import spawn, EOF

import sys
from sys import exit as sys_exit
from os import remove, path
from os import environ, system
from os.path import isfile
from urllib import URLopener
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT

pack_name = 'dapp-privatix.deb'
dwnld_url = 'http://art.privatix.net/{}'.format(pack_name)
dst_path = '{}'.format(pack_name)
# inst_pack = 'sudo apt update && sudo dpkg -i {} || sudo apt-get install -f -y'.format(dst_path)
# inst_pack = 'sudo sh -c \"apt update && dpkg -i {} || apt-get install -f -y\"'.format(dst_path)
# del_pack = 'sudo sh -c \"apt purge {} -y\"'.format((dst_path.split('.'))[0])
user_file = '/etc/sudoers.d/{}'

check_sudo = 'dpkg -l | grep sudo'
install_sudo = 'su -c \'apt install sudo\''


class Preparation():
    def __init__(self, log, bar, page_t,sys_pswd,sys_call_meth):
        self.inst_pack = 'sudo sh -c \"apt update && dpkg -i {} || ' \
                    'apt-get install -f -y\"'.format(dst_path)
        self.del_pack = 'sudo sh -c \"apt purge {} -y\"'.format(
            (dst_path.split('.'))[0])

        self.logging = log
        self.statusBar = bar
        self.pageText = page_t
        self.sys_pswd = sys_pswd
        self.sys_call_meth = sys_call_meth

    def sys_call(self, cmd):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()

        self.logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            self.logging.error(
                'Trouble when call: {}. Result: {}'.format(cmd, resp[1]))
            return False
        return resp[0]

    def ubn(self):
        self.logging.debug('Ubuntu')
        user = self.sys_call('whoami').replace('\n', '')
        f_path = user_file.format(user)

        if not isfile(f_path):

            cmd = 'sudo sh -c "echo \'{} ALL=(ALL:ALL) NOPASSWD:ALL\' >> {}"'.format(user, f_path)

            self.logging.debug('CMD: {}'.format(cmd))
            self.logging.info('Create sudoers file.')
            if self.sys_call_meth(cmd, self.sys_pswd):
                return False, 'Problem with {}'.format(cmd)
            return True, ''
        return True, 'File exist'

    def deb(self):
        self.logging.debug('Debian')
        sudo = self.sys_call(check_sudo)
        if not sudo:
            self.logging.info('Install sudo.\n')
            system(install_sudo)

        user = self.sys_call('whoami').replace('\n', '')
        f_path = user_file.format(user)

        if not isfile(f_path):
            cmd = 'su -c "echo \'{0} ALL=(ALL:ALL) NOPASSWD:ALL\' >> {1}"'.format(user,f_path)
            self.logging.info('Begin to create sudoers file.')
            self.logging.debug('CMD: {}'.format(cmd))
            if self.sys_call_meth(cmd, self.sys_pswd):
                return False, 'Problem with {}'.format(cmd)
            self.logging.debug('File sudoers created')
            return True, ''
        self.logging.debug('File sudoers exist')
        return True, 'File exist'

    def __check_dist(self):
        dist_name, ver, name_ver = linux_distribution()
        task = dict(ubuntu=self.ubn,
                    debian=self.deb
                    )
        self.logging.debug('OS Dist: {}'.format((dist_name, ver, name_ver)))
        dist_task = task.get(dist_name.lower(), False)
        if dist_task:
            return True, dist_task
        else:
            self.logging.info('You OS is not support yet.')
            return False, 'You OS is not support yet.'

    def __dwnld_pack(self):
        self.pageText.setHtml(
            'Preparing to download and install the package {}<br>'
            'Please wait'.format(pack_name))
        bar_act = self.statusBar(on=True)
        try:
            f = open(dst_path, "wb")
            self.logging.info("Downloading {}".format(dst_path))
            response = requests.get(dwnld_url, stream=True)
            total_length = response.headers.get('content-length')

            if total_length is None:  # no content length header
                f.write(response.content)
            else:
                dl = 0
                total_length = int(total_length)
                for data in response.iter_content(chunk_size=4096):
                    dl += len(data)
                    f.write(data)
                    done = int(100 * dl / total_length)
                    bar_act("value", done)


            self.statusBar()

            self.logging.info('The installation was successful done.\n'
                              'Run: sudo /opt/privatix/initializer/initializer.py')
            return True, ''
        except BaseException as dwnlexp:
            self.logging.error('Dwnld pack: {}'.format(dwnlexp))
            return False, 'Trouble with download pack'

    def preparation(self):
        """ Check & create sudoers file,
        download pack,return full cmd for install"""
        res = self.__check_dist()
        if not res[0]:
            return res
        res[1]()
        self.logging.debug('__check_dist done')
        res = self.__dwnld_pack()
        if not res[0]:
            return res
        self.logging.debug('__dwnld_pack done')

        return True, self.inst_pack
