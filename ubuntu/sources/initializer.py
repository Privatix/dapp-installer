#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
        Initializer on pure Python 2.7

        Version 0.2.2

        mode:
    python initializer.py  -h                              get help information
    python initializer.py                                  start full install
    python initializer.py --build                          create cmd for dapp
    python initializer.py --vpn start/stop/restart/status  control vpn servise
    python initializer.py --comm start/stop/restart/status control common servise
    python initializer.py --mass start/stop/restart/status control common + vpn servise
    python initializer.py --test                           start in test mode
    python initializer.py --no-gui                         install without GUI
    python initializer.py --update-back                    update all contaiter without GUI
    python initializer.py --update-mass                    update all contaiter with GUI
    python initializer.py --update-gui                     update only GUI
    python initializer.py --link                           use another link for download.if not use, def link in main_conf[link_download]
    python initializer.py --branch                         use another branch than 'develop' for download. template https://raw.githubusercontent.com/Privatix/dappctrl/{ branch }/
"""

import sys
import random
import logging
import socket
from signal import SIGINT, signal, pause
from contextlib import closing
from re import search, sub, findall, compile, match, IGNORECASE
from codecs import open
from uuid import uuid1
from threading import Thread
from shutil import copyfile, rmtree
from json import load, dump
from time import time, sleep
from urllib import URLopener
from urllib2 import urlopen
from os.path import isfile, isdir, exists
from argparse import ArgumentParser
from ConfigParser import ConfigParser
from platform import linux_distribution
from subprocess import Popen, PIPE, STDOUT
from distutils.version import StrictVersion
from stat import S_IEXEC, S_IXUSR, S_IXGRP, S_IXOTH
from os import remove, mkdir, path, environ, stat, chmod, listdir, getcwd, \
    system

"""
Exit code:
    1 - Problem with get or upgrade systemd ver
    2 - If version of Ubuntu lower than 16
    3 - If sysctl net.ipv4.ip_forward = 0 after sysctl -w net.ipv4.ip_forward=1
    4 - Problem when call system command from subprocess
    5 - Problem with operation R/W unit file 
    6 - Problem with operation download file 
    7 - Problem with operation R/W server.conf 
    8 - Default DB conf is empty, and no section 'DB' in dappctrl-test.config.json
    9 - Check the run of the database is negative
    10 - Problem with read dapp cmd from file
    11 - Problem NPM
    12 - Problem with run psql
    13 - Problem with ready Vpn 
    14 - Problem with ready Common 
    15 - The version npm is not satisfied. The user self reinstall
    16 - Problem with deleting one of the GUI pack
    17 - Exception in code, see logging
    18 - Exit Ctrl+C
    19 - OS not supported
    20 - In build mode,None in dappctrl_id
    21 - User from which was installing not root
    22 - Problem with read dappctrl.config.json
    23 - Problem with read dappvpn.config.json
    24 - Attempt to install gui on a system without gui
    25 - Problem with R/W dappctrlgui/settings.json
    26 - Problem with download dappctrlgui.tar.xz
    27 - The dappctrlgui package is not installed correctly.Missing file settings.json
    28 - Absent .dapp_cmd file
    29 - trouble when try install LXC
    30 - trouble when try check LXC contaiters 
    31 - found installed containers 
    32 - Problem with operation R/W LXC contaiter config or run sh or interfaces conf
    33 - User click [X].
    34 - User click Cancel.

    not save port 8000 from SessionServer and 9000 from PayServer if role is client
"""

main_conf = dict(
    log_path='/var/log/initializer.log',
    branch='develop',
    link_download='http://art.privatix.net/',
    mask=['/24', '255.255.255.0'],
    mark_final='/var/run/installer.pid',
    wait_mess='{}.Please wait until completed.\n It may take about 5-10 minutes.\n Do not turn it off.',
    tmp_var=None,
    del_pack='sudo apt purge {} -y',
    del_dirs='sudo rm -rf {}*',
    search_pack='sudo dpkg -l | grep {}',
    nspawn=False,
    openvpn_conf='etc/openvpn/config/server.conf',

    dappctrl_dev_conf_json='https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl-dev.config.json',
    dappctrl_conf_json='opt/privatix/config/dappctrl.config.local.json',
    dappvpn_conf_json='opt/privatix/config/dappvpn.config.json',

    back_nspwn=dict(
        addr='10.217.3.0',
        ports=dict(vpn=[], common=[],
                   mangmt=dict(vpn=None, common=None)),

        file_download=[
            'vpn.tar.xz',
            'common.tar.xz',
            'systemd-nspawn@vpn.service',
            'systemd-nspawn@common.service'],
        path_container='/var/lib/container/',
        path_vpn='vpn/',
        path_com='common/',
        path_unit='/lib/systemd/system/',
        openvpn_fields=[
            'server {} {}',
            'push "route {} {}"'
        ],
        openvpn_tun='dev {}',
        openvpn_port=['port 443', 'management'],

        unit_vpn='systemd-nspawn@vpn.service',
        unit_com='systemd-nspawn@common.service',
        unit_field={
            'ExecStopPost=/sbin/sysctl': False,

            'ExecStartPre=/sbin/iptables': 'ExecStartPre=/sbin/iptables -t nat -A POSTROUTING -s {} -o {} -j MASQUERADE\n',
            'ExecStopPost=/sbin/iptables': 'ExecStopPost=/sbin/iptables -t nat -D POSTROUTING -s {} -o {} -j MASQUERADE\n',
        },
        db_log='/var/lib/container/common/var/log/postgresql/postgresql-10-main.log',
        db_stat='database system is ready to accept connections',
        tor=dict(
            socks_port=9099,
            config='etc/tor/torrcprod',
            hostname='var/lib/tor/hidden_service/hostname'
        )

    ),
    back_lxc=dict(
        addr='10.0.4.0',
        common_octet='52',
        vpn_octet='51',
        ports=dict(vpn=[], common=[],
                   mangmt=dict(vpn=None, common=None)),

        install=[
            'sudo apt-add-repository ppa:ubuntu-lxc/stable -y',
            'sudo apt-get update',
            'sudo apt-get install lxc lxc-templates cgroup-bin bridge-utils debootstrap -y'
        ],
        path_container='/var/lib/lxc/',
        exist_contrs='lxc-ls -f',
        exist_contrs_ip=dict(),
        openvpn_port=['port 443', ''],

        bridge_cmd='ip addr show',
        bridge_conf='/etc/default/lxc-net',
        deff_lxc_cont_path='/var/lib/lxc/',
        lxc_cont_conf_name='config',
        lxc_cont_interfs='/rootfs/etc/network/interfaces',
        lxc_cont_fs_file='/rootfs/home/ubuntu/go/bin/dappctrl',
        kind_of_cont=dict(common='/rootfs/etc/postgresql/',
                          vpn='/rootfs/etc/openvpn/'),
        run_sh_path='/etc/init.d/',
        chmod_run_sh='sudo chmod +x {}',
        update_cont_conf='lxc-update-config -c {}',
        search_name='lxcbr',
        name_in_main_conf={
            'LXC_BRIDGE=': None,
            'LXC_ADDR=': None,
            'LXC_NETWORK=': None,
            'USE_LXC_BRIDGE=': None,
        },
        name_in_contnr_conf=[
            'lxc.network.ipv4',
            'lxc.uts.name',
            'hwaddr'
        ],

        file_download=[
            'dapp-common',
            'dapp-vpn',
            'lxc-common.tar.xz',
            'lxc-vpn.tar.xz'],
        path_vpn='vpn/',
        path_com='common/',

        db_log='/var/lib/lxc/{}rootfs/var/log/postgresql/postgresql-10-main.log',
        db_stat='database system is ready to accept connections',
        db_conf_path='/var/lib/lxc/{}rootfs/etc/postgresql/10/main/',
    ),

    build={
        'cmd': '/opt/privatix/initializer/dappinst -dappvpnconftpl=\'{0}\' -dappvpnconf={1} -connstr=\"{3}\" -template={4} -agent=true\n'
               '/opt/privatix/initializer/dappinst -dappvpnconftpl=\'{0}\' -dappvpnconf={2} -connstr=\"{3}\" -template={4} -agent=false',
        'cmd_path': '.dapp_cmd',

        'db_conf': {
            "dbname": "dappctrl",
            "sslmode": "disable",
            "user": "postgres",
            "host": "localhost",
            "port": "5433"
        },

        'dappvpnconf_path': '/var/lib/container/vpn/opt/privatix/config/dappvpn.config.json',
        'dappconconf_path': '/var/lib/container/common/opt/privatix/config/dappvpn.config.json',
        'conf_link': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/dappctrl.config.json',
        'templ': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/svc/dappvpn/dappvpn.config.json',
        'dappctrl_id_raw': 'https://raw.githubusercontent.com/Privatix/dappctrl/{}/data/prod_data.sql',
        'field_name_id': '--templateid = ',
        'dappctrl_id': None,
    },

    gui={
        'gui_arch': 'dappctrlgui.tar.xz',
        'gui_path': '/opt/privatix/gui/',
        'link_dev_gui': 'dappctrlgui/',
        'icon_name': 'privatix-dappgui.desktop',
        'icon_sh': 'privatix-dappgui.sh',
        'icon_dir': '{}{}/Desktop/',
        'icon_tmpl_f_sh': '{}{}/{}',
        'icon_tmpl': {
            'Section': 'Desktop Entry',
            'Comment': 'First Internet Broadband Marketplace powered by P2P VPN Network on Blockchain',
            'Terminal': 'false',
            'Name': 'Privatix Dapp',
            'Exec': 'sh -c "sudo /opt/privatix/initializer/initializer.py --mass start && sudo npm start --prefix /opt/privatix/gui/{}"',
            'Type': 'Application',
            'Icon': '/opt/privatix/gui/{}icon_64.png',
        },

        'icon_prod': 'node_modules/dappctrlgui/',
        'dappctrlgui': '/opt/privatix/gui/node_modules/dappctrlgui/settings.json',

        'npm_tmp_f': 'tmp_nodesource',
        'npm_url': 'https://deb.nodesource.com/setup_9.x',
        'npm_tmp_file_call': 'sudo -E bash ',
        'npm_node': 'sudo apt-get install -y nodejs',

        'gui_inst': [
            'sudo chown -R root:root /opt/privatix/gui/',
            'sudo su -c \'cd /opt/privatix/gui && sudo npm install dappctrlgui\''
        ],
        'chown': 'sudo chown -R {0}:$(id -gn {0}) {1}',
        'version': {
            'npm': ['5.6', None, '0'],
            'nodejs': ['9.0', None, '0'],
            'node': ['9.0', None, '0'],
        },

    },

    test={
        'path': 'test_data.sql',
        'sql': 'https://raw.githubusercontent.com/Privatix/dappctrl/develop/data/test_data.sql',
        'cmd': 'psql -d dappctrl -h 127.0.0.1 -P 5433 -f {}'
    },

    dnsmasq={
        'conf': '/etc/NetworkManager/NetworkManager.conf',
        'section': ['main', 'dns', 'dnsmasq'],
        'disable': 'sudo sed -i \'s/^dns=dnsmasq/#&/\' /etc/NetworkManager/NetworkManager.conf && '
                   'sudo service network-manager restart',
    },
)


class Init:
    recursion = 0
    target = None  # may will be back,gui,both
    sysctl = False
    waiting = True
    in_args = None  # arguments with which the script was launched.
    dappctrl_role = None  # the role: agent|client.

    def __init__(self):
        self.uid_dict = dict(userid=str(uuid1()))
        self.mask = main_conf['mask']
        self.tmp_var = main_conf['tmp_var']
        self.fin_file = main_conf['mark_final']
        self.url_dwnld = main_conf['link_download']
        self.wait_mess = main_conf['wait_mess']
        self.p_dapctrl_conf = main_conf[
            'dappctrl_conf_json']  # ping [3000,8000,9000]
        self.p_dapctrl_dev_conf = main_conf['dappctrl_dev_conf_json'].format(
            main_conf['branch'])
        self.p_dapvpn_conf = main_conf['dappvpn_conf_json']
        self.ovpn_conf = main_conf['openvpn_conf']

        test = main_conf['test']
        self.test_path = test['path']
        self.test_sql = test['sql']
        self.test_cmd = test['cmd']

        bld = main_conf['build']
        self.db_conf = bld['db_conf']
        self.build_cmd = bld['cmd']

        self.dupp_raw_id = bld['dappctrl_id_raw'].format(main_conf['branch'])
        self.dappctrl_id = bld['dappctrl_id']

        self.field_name_id = bld['field_name_id']
        self.dupp_conf_url = bld['conf_link'].format(main_conf['branch'])
        self.dupp_vpn_templ = bld['templ'].format(main_conf['branch'])
        self.build_cmd_path = bld['cmd_path']

        gui = main_conf['gui']
        self.gui_arch = gui['gui_arch']
        self.gui_path = gui['gui_path']
        self.gui_version = gui['version']
        self.gui_icon_sh = gui['icon_sh']
        self.dappctrlgui = gui['dappctrlgui']
        self.gui_npm_url = gui['npm_url']
        self.gui_dev_link = gui['link_dev_gui']
        self.gui_npm_node = gui['npm_node']
        self.gui_icon_name = gui['icon_name']
        self.gui_icon_path = gui['icon_dir']
        self.gui_icon_tmpl = gui['icon_tmpl']
        self.gui_icon_prod = gui['icon_prod']
        self.gui_installer = gui['gui_inst']
        self.gui_npm_tmp_f = gui['npm_tmp_f']
        self.gui_icon_chown = gui['chown']
        self.gui_icon_path_sh = gui['icon_tmpl_f_sh']
        self.gui_npm_cmd_call = gui['npm_tmp_file_call']

        dnsmasq = main_conf['dnsmasq']
        self.dns_conf = dnsmasq['conf']
        self.dns_sect = dnsmasq['section']
        self.dns_disable = dnsmasq['disable']

    def re_init(self):
        self.__init__()

    def __init_back(self, back):
        self.addr = back['addr']
        self.p_contr = back['path_container']
        self.f_dwnld = back['file_download']
        self.path_vpn = back['path_vpn']
        self.path_com = back['path_com']
        self.ovpn_port = back['openvpn_port']

        self.use_ports = back['ports']  # store all ports for monitor

        self.db_log = back['db_log']
        self.db_stat = back['db_stat']
        self.p_unpck = dict(
            vpn=[self.path_vpn, '0.0.0.0'],
            common=[self.path_com, '0.0.0.0']
        )
        tor = back['tor']
        self.tor_socks_port = tor['socks_port']
        self.tor_config = tor['config']
        self.tor_hostname_config = tor['hostname']

    def _init_nspwn(self):
        back = main_conf['back_nspwn']
        self.__init_back(back)
        self.ovpn_tun = back['openvpn_tun']
        self.ovpn_fields = back['openvpn_fields']
        self.unit_dest = back['path_unit']
        self.unit_f_com = back['unit_com']
        self.unit_f_vpn = back['unit_vpn']
        self.unit_params = back['unit_field']

    def _init_lxc(self):
        back = main_conf['back_lxc']
        self.__init_back(back)
        self.db_conf_path = back['db_conf_path']
        self.lxc_install = back['install']
        self.exist_contrs = back['exist_contrs']
        self.lxc_contrs = back['exist_contrs_ip']
        self.bridge_cmd = back['bridge_cmd']
        self.bridge_conf = back['bridge_conf']
        self.deff_lxc_cont_path = back['deff_lxc_cont_path']
        self.lxc_cont_conf_name = back['lxc_cont_conf_name']
        self.lxc_cont_interfs = back['lxc_cont_interfs']
        self.lxc_cont_fs_file = back['lxc_cont_fs_file']
        self.kind_of_cont = back['kind_of_cont']
        self.run_sh_path = back['run_sh_path']
        self.chmod_run_sh = back['chmod_run_sh']
        self.update_cont_conf = back['update_cont_conf']
        self.search_name = back['search_name']
        self.name_in_main_conf = back['name_in_main_conf']
        self.name_in_contnr_conf = back['name_in_contnr_conf']
        self.p_unpck['vpn'] = [self.path_vpn, back['vpn_octet']]
        self.p_unpck['common'] = [self.path_com, back['common_octet']]
        self.def_comm_addr = self.addr.split('.')
        self.def_comm_addr[-1] = back['common_octet']
        self.def_comm_addr = '.'.join(self.def_comm_addr)

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
    def wait_decor(self):
        def wrap(obj, args=None):
            logging.debug('Wait decor args: {}.'.format(args))
            st = Thread(target=Init.long_waiting)
            st.daemon = True
            st.start()
            if args:
                res = self(obj, args)
            else:
                res = self(obj)
            Init.waiting = False
            sleep(0.5)
            return res

        return wrap


class CommonCMD(Init):
    def __init__(self):
        Init.__init__(self)

    def _sysctl(self):
        """ Return True if ip_forward=1 by default,
        and False if installed by script """
        cmd = 'sudo /sbin/sysctl net.ipv4.ip_forward'
        res = self._sys_call(cmd).decode()
        param = int(res.split(' = ')[1])

        if not param:
            if self.recursion < 1:
                logging.debug('Change net.ipv4.ip_forward from 0 to 1')

                cmd = 'sudo /sbin/sysctl -w net.ipv4.ip_forward=1'
                self._sys_call(cmd)
                sleep(0.5)
                self.recursion += 1
                self._sysctl()
                return False
            else:
                logging.error('sysctl net.ipv4.ip_forward didnt change to 1')
                sys.exit(3)
        return True

    def _reletive_path(self, name):
        dirname = path.dirname(path.abspath(__file__))
        logging.debug('Reletive path: {}'.format(dirname))
        return path.join(dirname, name)

    def signal_handler(self, sign, frm):
        logging.info('You pressed Ctrl+C!')
        self._rolback(code=18)
        pause()

    def _clear_dir(self, p):
        logging.debug('Clear dir: {}'.format(p))
        cmd = main_conf['del_dirs'].format(p)
        return self._sys_call(cmd)

    def _rolback(self, code):
        # Rolback net.ipv4.ip_forward and clear store
        logging.debug('Rolback. sysctl: {}, code: {}'.format(self.sysctl, code))

        if not self.old_vers and not self.sysctl:
            logging.debug('Rolback ip_forward')
            cmd = 'sudo /sbin/sysctl -w net.ipv4.ip_forward=0'
            self._sys_call(cmd)

        self._clear_dir(self.p_contr)
        return self._clear_dir(self.gui_path)

    def service(self, srv, status, port=None, reverse=False):
        logging.debug(
            'Service:{}, port:{}, status:{}, reverse:{}'.format(srv, port,
                                                                status,
                                                                reverse))
        tmpl = ['sudo systemctl {} {} && sleep 0.5']
        rmpl_rest = ['sudo systemctl stop {1} && sleep 0.5',
                     'sudo systemctl start {1} && sleep 0.5']
        rmpl_stat = ['sudo systemctl is-active {1}']

        scroll = {'start': tmpl, 'stop': tmpl,
                  'restart': rmpl_rest, 'status': rmpl_stat}
        unit_serv = {'vpn': self.unit_f_vpn, 'comm': self.unit_f_com}

        if status not in scroll.keys():
            logging.error('Status {} not suitable for service {}. '
                          'Status must be one from {}'.format(
                status, srv, scroll.keys())
            )
            return None

        raw_res = list()
        for cmd in scroll[status]:
            cmd = cmd.format(status, unit_serv[srv])
            res = self._sys_call(cmd, rolback=False)

            if not port:
                continue

            if status == 'status':
                if res == 'active\n':
                    if reverse:
                        return self._checker_port(port=port, status='stop')
                    else:
                        return self._checker_port(port=port)
                else:
                    return False

            if status == 'restart':
                check_stat = 'start' if 'start' in cmd else 'stop'
            else:
                check_stat = status

            if 'failed' in res or not self._checker_port(port=port,
                                                         status=check_stat):
                return False
            raw_res.append(True)

        if not port:
            return None
        return all(raw_res)

    @Init.wait_decor
    def clear_contr(self, pass_check=False):
        # Stop container.Check it if pass_check True.Clear conteiner path
        if pass_check:
            logging.info('\n\n   --- Attention! ---\n'
                         ' During installation a failure occurred'
                         ' or you pressed Ctrl+C\n'
                         ' All installed will be removed and returned to'
                         ' the initial state.\n Wait for the end!\n '
                         ' And try again.\n   ------------------\n')
        self.service('vpn', 'stop', self.use_ports['vpn'])
        self.service('comm', 'stop', self.use_ports['common'])
        sleep(3)

        if pass_check or not self.service('vpn', 'status',
                                          self.use_ports['vpn'],
                                          True) and \
                not self.service('comm', 'status',
                                 self.use_ports['common'], True):
            self._clear_dir(self.p_contr)
            return True

    def file_rw(self, p, w=False, data=None, log=None, json_r=False):
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

    def stop_services(self):
        logging.debug('Stop all containers')
        self.service('vpn', 'stop')
        self.service('comm', 'stop')

    def run_service(self, comm=False, restart=False, nspwn=True):

        if comm:
            if nspwn:
                if restart:
                    logging.info('Restart common service')
                    self.service('comm', 'stop')
                    # self._sys_call(
                    #     'sudo systemctl stop {}'.format(self.unit_f_com))
                else:
                    logging.info('Run common service')
                    self._sys_call('sudo systemctl daemon-reload')
                    sleep(2)
                    self._sys_call(
                        'sudo systemctl enable {}'.format(self.unit_f_com))
                sleep(2)
                self.service('comm', 'start')
                # self._sys_call(
                #     'sudo systemctl start {}'.format(self.unit_f_com))
            else:
                if restart:
                    logging.info('Restart common service')
                    self._sys_call('sudo service dapp-common stop')
                self._sys_call('sudo service dapp-common start')

        else:
            if nspwn:
                if restart:
                    logging.info('Restart vpn service')
                    self.service('vpn', 'stop')
                    # self._sys_call(
                    #     'sudo systemctl stop {}'.format(self.unit_f_vpn))
                else:
                    logging.info('Run vpn service')
                    self._sys_call(
                        'sudo systemctl enable {}'.format(self.unit_f_vpn))
                sleep(2)
                self.service('vpn', 'start')

                # self._sys_call(
                #     'sudo systemctl start {}'.format(self.unit_f_vpn))
            else:
                if restart:
                    logging.info('Restart vpn service')
                    self._sys_call('sudo service dapp-vpn stop')

                self._sys_call('sudo service dapp-vpn start')

    def _sys_call(self, cmd, rolback=True, s_exit=4):
        resp = Popen(cmd, shell=True, stdout=PIPE,
                     stderr=STDOUT).communicate()
        logging.debug('Sys call cmd: {}. Stdout: {}'.format(cmd, resp))
        if resp[1]:
            logging.debug('Error in sys call: {}'.format(resp[1]))
            if rolback:
                self._rolback(s_exit)
            else:
                return False

        elif 'The following packages have unmet dependencies:' in resp[0]:
            if rolback:
                self._rolback(s_exit)
            return False
        return resp[0]

    def _disable_dns(self):
        logging.debug('Disable dnsmasq')
        if isfile(self.dns_conf):
            logging.debug('dnsmasq conf exist')
            cfg = ConfigParser()
            cfg.read(self.dns_conf)
            if cfg.has_option(self.dns_sect[0], self.dns_sect[1]) and \
                            cfg.get(self.dns_sect[0], self.dns_sect[1]) == \
                            self.dns_sect[2]:
                logging.debug('Section {}={} found.'.format(self.dns_sect[1],
                                                            self.dns_sect[
                                                                2]))

                logging.debug('Disable dnsmasq !')
                self._sys_call(self.dns_disable, rolback=False)
            else:
                logging.debug(
                    'dnsmasq conf has not {}'.format(self.dns_sect[0:2]))
        else:
            logging.debug('dnsmasq conf not exist')

    def _ping_port(self, port, host='0.0.0.0', verb=False):
        '''open -> True  close -> False'''

        with closing(
                socket.socket(socket.AF_INET, socket.SOCK_STREAM)) as sock:
            if sock.connect_ex((host, int(port))) == 0:
                if verb:
                    logging.info("Port {} is open".format(port))
                else:
                    logging.debug("Port {} is open".format(port))
                return True
            else:
                if verb:
                    logging.info("Port {} is not available".format(port))
                else:
                    logging.debug("Port {} is not available".format(port))
                return False

    def _cycle_ask(self, h, p, status, verb=False):
        logging.debug('Ask port: {}, status: {}'.format(p, status))
        ts = time()
        tw = 350

        if status == 'stop':
            logging.debug('Stop mode')
            while True:
                if not self._ping_port(port=p, host=h, verb=verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)
        else:
            logging.debug('Start mode')
            while True:
                if self._ping_port(port=p, host=h, verb=verb):
                    return True
                if time() - ts > tw:
                    return False
                sleep(2)

    def _checker_port(self, port, host='0.0.0.0', status='start',
                      verb=False):
        logging.debug('Checker: {}'.format(status))
        if not port:
            return None
        if isinstance(port, (list, set)):
            resp = list()
            for p in port:
                resp.append(self._cycle_ask(host, p, status, verb))
            return True if all(resp) else False
        else:
            return self._cycle_ask(host, port, status, verb)

    def __all_use_ports(self, d):
        for k, v in d.iteritems():
            if v is None:
                continue
            elif isinstance(v, dict):
                self.__all_use_ports(v)
            elif isinstance(v, list):
                self.tmp_var += map(int, v)
            else:
                self.tmp_var.append(int(v))

    def check_port(self, port, auto=False):
        '''_ping_port: open -> True  close -> False'''

        if self._ping_port(port=port):
            while True:

                if auto:
                    port = str(int(port)+1)
                else:
                    logging.info("\nPort: {} is busy or wrong.\n"
                                 "Select a different port,in range 1 - 65535.".format(
                        port))
                    port = raw_input('>')
                try:
                    self.tmp_var = []
                    self.__all_use_ports(self.use_ports)
                    if int(port) in range(65535)[1:] and not self._ping_port(
                            port=port) and int(port) not in self.tmp_var:
                        break
                except BaseException as bexpm:
                    logging.error('Check port: {}'.format(bexpm))

        return port

    @Init.wait_decor
    def _wait_up(self):
        logging.info(self.wait_mess.format('Run services'))

        logging.debug('Wait when up all ports: {}'.format(self.use_ports))

        # check common
        if not self._checker_port(
                host=self.p_unpck['common'][1],
                port=self.use_ports['common'],
                verb=True):
            logging.info('Restart Common')
            self.run_service(comm=True, restart=True)
            if not self._checker_port(
                    host=self.p_unpck['common'][1],
                    port=self.use_ports['common'],
                    verb=True):
                logging.error('Common is not ready')
                self._rolback(22)
                return False, 'Common is not ready'

        # check vpn
        if not self._checker_port(
                host=self.p_unpck['vpn'][1],
                port=self.use_ports['vpn'],
                verb=True):
            logging.info('Restart VPN')
            self.run_service(comm=False, restart=True)
            if not self._checker_port(
                    host=self.p_unpck['vpn'][1],
                    port=self.use_ports['vpn'],
                    verb=True):
                logging.error('VPN is not ready')
                self._rolback(22)
                return False, 'VPN is not ready'

        return True, ''

    def _finalizer(self, rw=None):

        if not isfile(self.fin_file):
            cmd = 'sudo touch {0} && sudo chmod 677 {0}'.format(
                self.fin_file)
            logging.debug('Create PID file: {}'.format(self.fin_file))
            self._sys_call(cmd=cmd)
            # self.file_rw(p=self.fin_file, w=True, log='First start')
            return True, 'Pid file created'

        if rw:
            self.file_rw(p=self.fin_file, w=True, data=self.use_ports,
                         log='Finalizer.Write port info', json_r=True)
            return True, 'Pid file rewrited'

        mark = self.file_rw(p=self.fin_file)
        logging.debug('Start marker: {}'.format(mark))
        if not mark:
            logging.info('First start')
            return True, 'First start'

        mess = 'Second start.' \
               'This is protection against restarting the program.' \
               'If you need to re-run the script, ' \
               'you need to delete the file {}'.format(self.fin_file)

        logging.info(mess)
        return False, mess

    def __exclude_port(self, tmp_store):
        logging.debug('Exclude port mode.')
        # only if agent is client !
        by_key = ['PayServer', 'SOMCServer']  # 9000,5555
        for i in by_key:
            if tmp_store.get(i):
                logging.debug('Exclude: {}'.format(i))
                del tmp_store[i]

        return tmp_store

    def conf_dappctrl_json(self, old_vers=False):
        """Check ip addr, free ports and replace it in
        common dappctrl.config.local.json"""
        logging.debug('Check IP, Port in common dappctrl.local.json')
        # search_keys = ['AgentServer', 'PayAddress', 'PayServer',
        #                'SessionServer']
        pay_port = dict(old=None, new=None)
        p = self.p_contr + self.path_com
        if old_vers:
            p += 'rootfs/'
        p += self.p_dapctrl_conf

        # Read dappctrl.config.local.json
        data = self.file_rw(p=p, json_r=True, log='Read dappctrl conf')
        if not data:
            self._rolback(22)
            return False

        if old_vers:
            # LXC mode
            r = data.get('DB').get('Conn')
            if r:
                r.update({"host": self.p_unpck['common'][1]})

        # Check and change self ip and port for PayAddress
        my_ip = urlopen(url='http://icanhazip.com').read().replace('\n', '')
        logging.debug('Found IP: {}. Write it.'.format(my_ip))

        raw = data['PayAddress'].split(':')

        raw[1] = '//{}'.format(my_ip)

        delim = '/'
        rout = raw[-1].split(delim)
        pay_port['old'] = rout[0]
        pay_port['new'] = self.check_port(pay_port['old'], True)
        rout[0] = pay_port['new']
        raw[-1] = delim.join(rout)

        data['PayAddress'] = ':'.join(raw)

        # change role: agent, client
        data['Role'] = self.dappctrl_role

        # Search ports in conf and store it to main_conf['ports']
        tmp_store = dict()
        for k, v in data.iteritems():
            if isinstance(v, dict) and v.get('Addr'):
                delim = ':'
                raw_row = v['Addr'].split(delim)
                port = raw_row[-1]
                logging.debug('Key: {} port: {}, Check it.'.format(k, port))

                if k == 'PayServer':
                    # default Addr is 0.0.0.0:9000
                    # ping only when role agent
                    raw_row[-1] = pay_port['new']
                    # if self.dappctrl_role == 'agent':
                    tmp_store[k] = pay_port['new']

                else:
                    if old_vers and raw_row[0] == self.def_comm_addr:
                        # LXC mode
                        raw_row[0] = self.p_unpck['common'][1]

                    port = self.check_port(port, True)
                    raw_row[-1] = port

                    tmp_store[k] = port
                    if k == 'UI':
                        # default Addr is localhost:8888
                        self.wsEndpoint = port
                        self.use_ports['wsEndpoint'] = port

                    elif k == 'SessionServer':
                        # default Addr is localhost:8000
                        self.sessServPort = port

                    # elif k == 'SOMCServer' and self.dappctrl_role == 'client':
                        # default Addr is localhost:5555
                        # ping only when role agent
                        # self.use_ports['common'].remove(port)

                data[k]['Addr'] = delim.join(raw_row)

        # add uid key in conf
        logging.debug('Add userid on dappctrl.config.local.json')
        if data.get('Report'):
            data['Report'].update(self.uid_dict)
        else:
            data['Report'] = self.uid_dict

        # Rewrite dappctrl.config.local.json
        self.file_rw(p=p, w=True, json_r=True, data=data,
                     log='Rewrite conf')

        if self.dappctrl_role == 'client':
            tmp_store = self.__exclude_port(tmp_store)

        self.use_ports['common'] = [v for k, v in tmp_store.items()]

        return True


class Tor(CommonCMD):
    def __init__(self):
        CommonCMD.__init__(self)

    def check_tor_port(self):
        logging.info('TOR. Check config')
        self.tor_socks_port = self.check_port(port=self.tor_socks_port,
                                     auto=True)
        full_comm_p = self.p_contr + self.path_com
        data = self.file_rw(p=full_comm_p + self.p_dapctrl_conf,
                            json_r=True,
                            log='TOR. Read dappctrl conf')


        somc_serv = data.get('SOMCServer')
        if somc_serv:
            somc_serv_port = somc_serv['Addr'].split(':')[1]
            serv_port = '80 127.0.0.1:{}\n'.format(somc_serv_port)
            logging.debug('TOR. HiddenServicePort: {}'.format(serv_port))
            data = self.file_rw(p=full_comm_p + self.tor_config,
                                log='TOR. Read tor conf.')
            if not data:
                raise BaseException('Tor config are absent!')

            search_line = {
                'SocksPort': '{}\n'.format(self.tor_socks_port),
                'HiddenServicePort': serv_port
            }

            for row in data:
                for k, v in search_line.items():
                    if k in row:
                        indx = data.index(row)
                        data[indx] = '{} {}'.format(k, v)

            self.file_rw(p=full_comm_p + self.tor_config,
                         w=True,
                         data=data,
                         log='TOR. Write tor conf.')

        else:
            mess = 'TOR. On dappctrl.config.json absent SOMCServer field'
            logging.error(mess)
            raise BaseException(mess)

    def get_onion_key(self):
        logging.debug('TOR. Get onion key')
        hostname_config = self.p_contr + self.path_com + self.tor_hostname_config
        onion_key = self.file_rw(
            p=hostname_config,
            log='TOR. Read hostname conf')
        logging.debug('TOR. Onion key: {}'.format(onion_key))

        data = self.file_rw(
            p=self.p_contr + self.path_com + self.p_dapctrl_conf,
            json_r=True,
            log='TOR. Read dappctrl conf')

        data.update(dict(
            TorHostname=onion_key[0].replace('\n', '')
        ))

        self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                     w=True,
                     json_r=True,
                     data=data,
                     log='TOR. Write add TorHostname to dappctrl conf.')

    def set_socks_list(self):
        logging.debug('TOR. Add TorSocksListener.')
        data = self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                            json_r=True,
                            log='TOR. Read dappctrl conf')
        data.update(dict(
            TorSocksListener=self.tor_socks_port
        ))

        self.file_rw(p=self.p_contr + self.path_com + self.p_dapctrl_conf,
                     json_r=True,
                     w=True,
                     data=data,
                     log='TOR. Write dappctrl conf')


class DB(Tor):
    '''This class provides a check if the database is started from its logs'''

    def __init__(self):
        Tor.__init__(self)

    @Init.wait_decor
    def _check_db_run(self, code):
        # wait 't_wait' sec until the DB starts, if not started, exit.

        t_start = time()
        t_wait = 300
        mark = True
        logging.info('Waiting for the launch of the DB.')
        while mark:
            logging.debug('Wait.')
            raw = self.file_rw(p=self.db_log,
                               log='Read DB log')
            for i in raw:
                logging.debug('DB : {}'.format(i))

                if self.db_stat in i:
                    logging.info('DB was run.')
                    mark = False
                    return True
            if time() - t_start > t_wait:
                logging.error(
                    'DB after {} sec does not run.'.format(t_wait))
                logging.debug('Data base log: \n  {}'.format(raw))
                self._rolback(code)
                return False
            sleep(5)

    def _clear_db_log(self):
        logging.debug('Clear DB log')
        self._sys_call(cmd='sudo chmod 647 {}'.format(self.db_log))
        self._sys_call(cmd='sudo echo \'\'>{}'.format(self.db_log))


class Params(DB):
    """ This class provide check
    sysctl, iptables, port, ip"""

    def __init__(self):
        DB.__init__(self)

    def _rw_unit_file(self, ip, intfs, code):
        logging.debug('Preparation unit file: {},{}'.format(ip, intfs))
        addr = ip + main_conf['mask'][0]
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(p=self.p_contr + self.unit_f_vpn)
            logging.debug('Read {}'.format(self.unit_f_vpn))
            # replace all search fields
            for row in tmp_data:

                for param in self.unit_params.keys():
                    if param in row:
                        indx = tmp_data.index(row)

                        if self.unit_params[param]:
                            tmp_data[indx] = self.unit_params[param].format(
                                addr,
                                intfs)
                        else:
                            if self.sysctl:
                                tmp_data[indx] = ''

            # rewrite unit file
            logging.debug('Rewrite {}'.format(self.unit_f_vpn))
            self.file_rw(p=self.p_contr + self.unit_f_vpn, w=True,
                         data=tmp_data)
            del tmp_data

            # move unit files
            logging.debug('Move units.')
            system('sudo cp {} {}'.format(self.p_contr + self.unit_f_vpn,
                                          self.unit_dest + self.unit_f_vpn))
            system('sudo cp {} {}'.format(self.p_contr + self.unit_f_com,
                                          self.unit_dest + self.unit_f_com))
        except BaseException as f_rw:
            logging.error('R/W unit file: {}'.format(f_rw))
            self._rolback(code)

    def _check_dapp_conf(self):
        for servs, port in self.use_ports['mangmt'].iteritems():

            logging.debug('Dapp {} conf. Port: {}'.format(servs, port))
            if servs == 'vpn':
                p = self.p_contr + self.path_vpn + self.p_dapvpn_conf

            elif servs == 'common':
                p = self.p_contr + self.path_com + self.p_dapvpn_conf

            raw_data = self.file_rw(p=p,
                                    log='Check dapp {} conf'.format(servs),
                                    json_r=True)
            if not raw_data:
                self._rolback(23)
            # "localhost:7505"
            logging.debug('dapp {} conf: {}'.format(servs, raw_data))
            delim = ':'
            raw_tmp = raw_data['Monitor']['Addr'].split(delim)
            raw_tmp[-1] = str(port)
            raw_data['Monitor']['Addr'] = delim.join(raw_tmp)
            logging.debug(
                'Monitor Addr: {}.'.format(raw_data['Monitor']['Addr']))

            if hasattr(self, 'sessServPort'):
                delim = ':'
                raw_tmp = raw_data['Connector']['Addr'].split(delim)
                raw_tmp[-1] = str(self.sessServPort)
                raw_data['Connector']['Addr'] = delim.join(raw_tmp)
                logging.debug(
                    'Connector Addr: {}.'.format(raw_data['Connector']['Addr']))

            self.file_rw(p=p,
                         log='Rewrite {} conf'.format(servs),
                         data=raw_data,
                         w=True,
                         json_r=True)

    def _run_dapp_cmd(self):
        # generate two conf in path:
        #  /var/lib/container/vpn/opt/privatix/config/dappvpn.config.json
        #  /var/lib/container/common/opt/privatix/config/dappvpn.config.json

        cmds = self.file_rw(
            p='/opt/privatix/initializer/.dapp_cmd',
            log='Read dapp cmd')
        logging.debug('Dupp cmds: {}'.format(cmds))

        if cmds:
            for cmd in cmds:
                cmd = 'sudo ' + cmd
                self._sys_call(cmd=cmd)
                sleep(1)
        else:
            logging.error('Have not {} file for further execution. '
                          'It is necessary to run the initializer '
                          'in build mode.'.format(self.build_cmd_path))
            self._rolback(10)

class Rdata(CommonCMD):
    ''' Class for download, unpack, clear data '''

    def __init__(self):
        CommonCMD.__init__(self)

    @Init.wait_decor
    def download(self, code=6):
        try:
            logging.info('Begin download files.')
            dev_url = ''
            if not isdir(self.p_contr):
                logging.debug('Create dir: {}'.format(self.p_contr))
                mkdir(self.p_contr)

            obj = URLopener()
            if hasattr(self, 'back_route'):
                dev_url = self.back_route + '/'
                logging.debug('Back dev rout: "{}"'.format(self.back_route))

            for f in self.f_dwnld:
                logging.info(
                    self.wait_mess.format('Start download {}'.format(f)))

                logging.debug(
                    'url_dwnld:{}, dev_url:{} ,f: {}'.format(self.url_dwnld,
                                                             dev_url, f))
                dwnld_url = self.url_dwnld + '/' + dev_url + f
                dwnld_url = dwnld_url.replace('///', '/')
                logging.debug(' - dwnld url: "{}"'.format(dwnld_url))
                obj.retrieve(dwnld_url, self.p_contr + f)
                sleep(0.1)
                logging.info('Download {} done.'.format(f))
            return True

        except BaseException as down:
            logging.error('Download: {}.'.format(down))
            self._rolback(code)
            return False

    @Init.wait_decor
    def unpacking(self):
        logging.info('Begin unpacking download files.')
        try:
            for f in self.f_dwnld:
                if '.tar.xz' == f[-7:]:
                    logging.info('Unpacking {}.'.format(f))

                    for k, v in self.p_unpck.items():
                        if k in f:
                            if not isdir(self.p_contr + v[0]):
                                mkdir(self.p_contr + v[0])
                            cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                                self.p_contr + f, self.p_contr + v[0])
                            self._sys_call(cmd)
                            logging.info('Unpacking {} done.'.format(f))

        except BaseException as expt_unpck:
            logging.error('Unpack: {}.'.format(expt_unpck))

    def clean(self):
        logging.info('Delete downloaded files.')

        for f in self.f_dwnld:
            logging.info('Delete {}'.format(f))
            remove(self.p_contr + f)


class GUI(CommonCMD):
    def __init__(self):
        CommonCMD.__init__(self)
        self.icon_task = None
        self.__init_icon_path()

    def __init_icon_path(self):
        if environ.get('SUDO_USER'):
            logging.debug('SUDO_USER')
            if self.__check_desctop_dir('/home/', environ['SUDO_USER']):
                self.gui_icon = self.gui_icon_path.format(
                    '/home/', environ['SUDO_USER']) + self.gui_icon_name

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['SUDO_USER'], self.gui_icon
                )
                self.icon_task = self.__create_icon
            else:
                self.gui_icon = self.gui_icon_path_sh.format(
                    '/home/', environ['SUDO_USER'], self.gui_icon_sh)

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['SUDO_USER'], self.gui_icon
                )
                self.icon_task = self.__create_icon_sh

        else:
            logging.debug('HOME')
            if self.__check_desctop_dir('', environ['HOME']):
                self.gui_icon = self.gui_icon_path.format(
                    '', environ['HOME']) + self.gui_icon_name

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['USER'], self.gui_icon
                )
                self.icon_task = self.__create_icon

            else:
                self.gui_icon = self.gui_icon_path_sh.format(
                    '', environ['HOME'], self.gui_icon_sh)

                self.chown_cmd = self.gui_icon_chown.format(
                    environ['USER'], self.gui_icon
                )
                self.icon_task = self.__create_icon_sh

    def _prepare_icon(self):
        mess = 'Icon of application was created by path<br>{}.<br>'.format(self.gui_icon)
        self.icon_task()
        logging.debug(mess)
        mess = mess.format(self.gui_icon)
        return mess

    def __check_desctop_dir(self, p, u):
        if not isdir(self.gui_icon_path.format(p, u)):
            logging.debug(
                '{} not exist'.format(self.gui_icon_path.format(p, u)))
            return False
        logging.debug('{} exist'.format(self.gui_icon_path.format(p, u)))
        return True

    def __create_icon_sh(self):
        logging.debug('Create file: {}'.format(self.gui_icon))

        logging.info('The directory needed to create the startup '
                     'icon file was not found.\n'
                     'After the installation is complete, to run the program\n'
                     'you will need to run the file "sudo {}".\n'
                     'Press enter to continue.'.format(self.gui_icon))

        raw_input('')
        with open(self.gui_icon, 'w') as icon:
            cmd = self.gui_icon_tmpl['Exec']

            icon.writelines(cmd)

        self.__icon_rights()

    def __icon_rights(self):
        logging.debug('Create {} file done'.format(self.gui_icon))

        chmod(self.gui_icon,
              stat(self.gui_icon).st_mode | S_IXUSR | S_IXGRP | S_IXOTH)
        logging.debug('Chmod file done')
        self._sys_call(self.chown_cmd)
        logging.debug('Chown file done')

    def __create_icon(self):
        config = ConfigParser()
        config.optionxform = str
        section = self.gui_icon_tmpl['Section']
        logging.debug('Create file: {}'.format(self.gui_icon))

        with open(self.gui_icon, 'w') as icon:
            config.add_section(section)
            [config.set(section, k, v) for k, v in
             self.gui_icon_tmpl.items()]
            config.write(icon)

        self.__icon_rights()

    @Init.wait_decor
    def __get_gui(self):
        if hasattr(self, 'gui_route'):
            try:

                dev_url = self.url_dwnld + self.gui_dev_link + self.gui_route + '/' + self.gui_arch
                self._sys_call(self.gui_installer[0], s_exit=11)
                logging.debug('Gui dev rout: "{}"'.format(dev_url))
                obj = URLopener()
                obj.retrieve(dev_url, self.gui_path + self.gui_arch)
                logging.info('Download {} done.'.format(self.gui_arch))
                logging.info('Begin unpacking download file.')

                cmd = 'tar xpf {} -C {} --numeric-owner'.format(
                    self.gui_path + self.gui_arch, self.gui_path)
                self._sys_call(cmd)
                logging.info('Unpacking {} done.'.format(self.gui_arch))

                cmd = 'cd / && sudo npm install --prefix /opt/privatix/gui && sudo chown -R root:root /opt/privatix/gui'
                self._sys_call(cmd)
                self.dappctrlgui = '/opt/privatix/gui/settings.json'

                self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl[
                    'Exec'].format('')
                self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl[
                    'Icon'].format('')

            except BaseException as down:
                logging.error('Download {}.'.format(down))
                self._rolback(26)
                return False, down

        else:
            self.gui_icon_tmpl['Exec'] = self.gui_icon_tmpl['Exec'].format(
                self.gui_icon_prod)

            self.gui_icon_tmpl['Icon'] = self.gui_icon_tmpl['Icon'].format(
                self.gui_icon_prod)
            for cmd in self.gui_installer:
                self._sys_call(cmd, s_exit=11)

        if not isfile(self.dappctrlgui):
            logging.error(
                'The dappctrlgui package is not installed correctly')
            self._rolback(27)
            return False, 'The dappctrlgui package is not installed correctly'

        return True, ''

    def __rewrite_config(self):
        """
        /opt/privatix/gui/node_modules/dappctrlgui/settings.json
        example data structure:
        {
            "firstStart": false,
            "accountCreated": true,
            "wsEndpoint": "ws://localhost:8888/ws",
            "gas": {
                "acceptOffering": 100000,
                "createOffering": 100000,
                "transfer": 100000
            },
            "network": "rinkeby"
        }
        """
        try:
            raw_data = self.file_rw(p=self.dappctrlgui,
                                    log='Read settings.json',
                                    json_r=True)
            delim = ':'
            raw_link = raw_data['wsEndpoint'].split(delim)
            raw_link[-1] = '{}/ws'.format(self.wsEndpoint)
            raw_data['wsEndpoint'] = delim.join(raw_link)

            # add uid key in conf
            logging.debug('Add userid on settings.json')

            if raw_data.get('bugsnag'):
                raw_data['bugsnag'].update(self.uid_dict)
            else:
                raw_data['bugsnag'] = self.uid_dict

            # Rewrite settings.json
            if not self.file_rw(p=self.dappctrlgui,
                                w=True,
                                data=raw_data,
                                json_r=True,
                                log='Rewrite settings.json'):
                raise BaseException(
                    '{} was not found.'.format(self.dappctrlgui))
            return True, 'R\W settings.json done'
        except BaseException as rwconf:
            logging.error('R\W settings.json: {}'.format(rwconf))
            self._rolback(25)
            return False, rwconf

    @Init.wait_decor
    def __get_npm(self):
        # install npm and nodejs
        logging.debug('Get NPM for GUI.')
        # npm_path = self._reletive_path(self.gui_npm_tmp_f)
        npm_path = self.gui_npm_tmp_f
        logging.debug('Npm path: {}'.format(npm_path))
        logging.debug('Npm url: {}'.format(self.gui_npm_url))
        if self.old_vers:
            logging.debug('Download node for lxc.')
            cmd = 'wget -O {} -q \'{}\''.format(npm_path, self.gui_npm_url)
            self._sys_call(cmd=cmd, s_exit=11)

        else:
            self.file_rw(
                p=npm_path,
                w=True,
                data=urlopen(self.gui_npm_url),
                log='Download nodesource'
            )

        cmd = self.gui_npm_cmd_call + npm_path
        self._sys_call(cmd=cmd, s_exit=11)

        cmd = self.gui_npm_node
        self._sys_call(cmd=cmd, s_exit=11)

    def __get_pack_ver(self):
        res = False
        for pack, v in self.gui_version.items():
            logging.info('Check {} version.'.format(pack))
            cmd = '{} -v'.format(pack)
            exist_pack = system(cmd)
            logging.debug('Exist pack result: {}'.format(exist_pack))
            if not exist_pack:
                # pack exist -> 0
                res = True
                raw = self._sys_call(cmd=cmd)
                ver = '.'.join(findall('\d+', raw)[0:2])

                if StrictVersion(ver) < StrictVersion(v[0]):
                    self.gui_version[pack][1] = True
                    self.gui_version[pack][2] = ver

            else:
                logging.info('{} not installed yet.'.format(pack))
        return res

    def search_gui_pack(self):
        self.__get_pack_ver()

        logging.debug('Check dependencies.')
        user_mess = [False, []]
        if any([x[1] for x in self.gui_version.values()]):
            logging.info('\n\nYou have installed obsolete packages.\n'
                         'To continue the installation, '
                         'you need to update the following packages:')
            for k, v in self.gui_version.items():
                if v[1]:
                    mess = ' - {} ver {}. Min requirements: {}<br>'.format(k, v[2],
                                                                   v[0])
                    logging.info(mess)
                    user_mess[0] = True
                    user_mess[1].append(mess)
        logging.debug('Mess. {}'.format(user_mess))

        return user_mess
        #
        # for k, v in self.gui_version.items():
        #     if v[1]:
        #         logging.info('Preparing for deletion '
        #                      '{} {}'.format(k, v[2]))
        #         cmd = main_conf['del_pack'].format(k)
        #         self._sys_call(cmd=cmd)
        #
        # if self.__get_pack_ver():
        #     logging.info('The problem with deleting one of the listed '
        #                  'packages. Try to delete in manual mode '
        #                  'and repeat the process again.')
        #     self._rolback(16)

    def install_gui(self, perm):
        logging.debug('Install GUI.')
        # self.search_gui_pack()
        self.__get_npm()
        res = self.__get_gui()
        if res[0]:
            icon_mess = self._prepare_icon()
            perm(cycle='gui')
            res = self.__rewrite_config()
            if res[0]:
                return True, icon_mess
            return res
        return res

    # def _clear_gui(self):
    #     logging.info('Clear GUI.')
    #     p = self.gui_path
    #     self._clear_dir(p)

    # def update_gui(self):
    #     logging.info('Update GUI.')
    #     self.wsEndpoint = self.use_ports.get('wsEndpoint')
    #     if not self.use_ports.get('wsEndpoint'):
    #         logging.info('You can not upgrade GUI before '
    #                      'you not complete the installation.')
    #         sys.exit()
    #
    #     self._clear_gui()
    #     self.__get_gui()


class Nspawn(Params):
    def __init__(self):
        self._init_nspwn()
        Params.__init__(self)

    def _rw_openvpn_conf(self, new_ip, new_tun, new_port, code):
        # rewrite in /var/lib/container/vpn/etc/openvpn/config/server.conf
        # two fields: server,push "route",  if ip =! default addr.
        logging.debug('Nspawn openvpn_conf')
        conf_file = "{}{}{}".format(self.p_contr,
                                    self.path_vpn,
                                    self.ovpn_conf)
        def_ip = self.addr
        def_mask = main_conf['mask'][1]
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:
                for field in self.ovpn_fields:
                    if field.format(def_ip, def_mask) in row:
                        indx = tmp_data.index(row)
                        tmp_data[indx] = field.format(new_ip,
                                                      def_mask) + '\n'

                if self.ovpn_tun.format('tun') in row:
                    logging.debug(
                        'Rewrite tun interface on: {}'.format(new_tun))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = self.ovpn_tun.format(new_tun) + '\n'

                elif self.ovpn_port[0] in row:
                    logging.debug('Rewrite port on: {}'.format(new_port))
                    indx = tmp_data.index(row)
                    tmp_data[indx] = 'port {}\n'.format(new_port)

                elif self.ovpn_port[1] in row:
                    # management 127.0.0.1 7505
                    indx = tmp_data.index(row)
                    delim = ' '
                    raw_row = row.split(delim)
                    port = int(raw_row[-1])
                    logging.debug('Raw port: {}'.format(port))

                    self.use_ports['mangmt']['vpn'] = self.check_port(port,
                                                                      True)

                    self.use_ports['mangmt']['common'] = self.check_port(
                        int(self.use_ports['mangmt']['vpn']) + 1, True)

                    raw_row[-1] = '{}\n'.format(
                        self.use_ports['mangmt']['vpn'])
                    tmp_data[indx] = delim.join(raw_row)
            logging.debug('--server.conf')

            # rewrite server.conf file
            if not self.file_rw(
                    p=conf_file,
                    w=True,
                    data=tmp_data,
                    log='Rewrite server.conf'
            ):
                self._rolback(7)

            del tmp_data

            logging.debug('server.conf done')
        except BaseException as f_rw:
            logging.error('R/W server.conf: {}'.format(f_rw))
            self._rolback(code)

    def __upgr_sysd(self, cmd):
        try:
            raw = self._sys_call('systemd --version')

            ver = raw.split('\n')[0].split(' ')[1]
            logging.debug('systemd --version: {}'.format(ver))

            if int(ver) < 229:
                logging.info('Upgrade systemd')

                raw = self._sys_call(cmd)

                if self.recursion < 1:
                    self.recursion += 1

                    logging.info('Install systemd')
                    logging.debug(self.__upgr_sysd(cmd))
                else:
                    raise BaseException(raw)
                logging.info('Upgrade systemd done')

            logging.info('Systemd version: {}'.format(ver))
            self.recursion = 0

        except BaseException as sysexp:
            logging.error('Get/upgrade systemd ver: {}'.format(sysexp))
            sys.exit(1)

    def _nsp_ubu_pack(self):
        logging.debug('Update')
        self._sys_call('sudo apt-get update')
        logging.debug('Install systemd-container')
        self._sys_call('sudo apt-get install systemd-container -y')
        self._sys_call('sudo apt-get install lshw -y')
        self._disable_dns()

    def _nsp_deb_pack(self):

        cmd = 'sudo sh -c "echo deb http://http.debian.net/debian jessie-backports main ' \
              '> /etc/apt/sources.list.d/jessie-backports.list"'
        logging.debug('Add jessie-backports.list')
        self._sys_call(cmd)
        logging.debug('Update')
        self._sys_call('sudo apt-get update')
        self._sys_call('sudo apt-get install lshw -y')
        self.__upgr_sysd(
            cmd='sudo apt-get -t jessie-backports install systemd -y')

        logging.debug('Install systemd-container')
        self._sys_call('sudo apt-get install systemd-container -y')


class LXC(DB):
    def __init__(self):
        self._init_lxc()
        DB.__init__(self)

    def conf_dappvpn_json(self):
        """Check addr in vpn dappvpn.config.json"""
        logging.debug('Check addr in vpn dappvpn.config.json')
        search_keys = ['Monitor', 'Connector']
        delim = ":"
        for cont_path in (self.path_com, self.path_vpn):

            p = self.p_contr + cont_path + 'rootfs/' + self.p_dapvpn_conf

            # Read dappctrl.config.local.json
            data = self.file_rw(p=p, json_r=True, log='Read dappvpn conf')
            if not data:
                self._rolback(22)

            serv_addr = data.get('Connector').get('Addr')
            if serv_addr:
                raw = serv_addr.split(delim)
                raw[0] = self.p_unpck['common'][1]
                data['Connector']['Addr'] = delim.join(raw)
            else:
                logging.error('Field Connector not exist')

            monit_addr = data.get('Monitor').get('Addr')
            if monit_addr:
                raw = monit_addr.split(delim)
                raw[0] = self.p_unpck['vpn'][1]
                data['Monitor']['Addr'] = delim.join(raw)
            else:
                logging.error('Field Monitor not exist')

            # Rewrite dappvpn.config.json
            self.file_rw(p=p, w=True, json_r=True, data=data,
                         log='Rewrite conf')

    def _check_cont_addr(self):
        # Check if 0,0,ххх,51 & 10,0,ххх,52
        # If not, increment 4-th octet and check again
        def increment_octet(octet, all_ip):
            if octet in all_ip:
                octet = random.randint(5, 254)
                increment_octet(octet, all_ip)
            return octet

        found_contrs_ip = [v.get('lxc.network.ipv4', '...').split('.')
                           for k, v in self.lxc_contrs.items()]

        found_contrs_ip = [int(ip[3])
                           for ip in found_contrs_ip
                           if ip and ip[3].isdigit()]

        for name, addr in self.p_unpck.items():
            octet = increment_octet(int(addr[1]), found_contrs_ip)
            self.p_unpck[name][1] = str(octet)

    def _rw_openvpn_conf(self, code):
        # rewrite in /var/lib/lxc/vpn/rootfs/etc/openvpn/config/server.conf
        # management field
        logging.debug('Lxc openvpn_conf')
        conf_file = "{}{}{}{}".format(self.p_contr,
                                      self.path_vpn,
                                      'rootfs/',
                                      self.ovpn_conf)
        try:
            # read a list of lines into data
            tmp_data = self.file_rw(
                p=conf_file,
                log='Read openvpn server.conf'
            )

            # replace all search fields
            for row in tmp_data:

                if 'management' in row:
                    logging.debug(
                        'Rewrite management: {}'.format(row))
                    indx = tmp_data.index(row)
                    row = row.split(' ')
                    row[1] = self.p_unpck['vpn'][1]

                    tmp_data[indx] = ' '.join(row)

            logging.debug('--server.conf')

            # rewrite server.conf file
            if not self.file_rw(
                    p=conf_file,
                    w=True,
                    data=tmp_data,
                    log='Rewrite server.conf'
            ):
                self._rolback(7)

            del tmp_data

            logging.debug('server.conf done')
        except BaseException as f_rw:
            logging.error('R/W server.conf: {}'.format(f_rw))
            self._rolback(code)

    @Init.wait_decor
    def __install_lxc(self):
        logging.debug('Install lxc')

        for cmd in self.lxc_install:
            if not self._sys_call(cmd=cmd, rolback=False):
                logging.error('Error when try: {}.'.format(cmd))
                sys.exit(29)

    def __change_mac(self, macs):
        mac = "00:16:3e:%02x:%02x:%02x" % (
            random.randint(0, 255),
            random.randint(0, 255),
            random.randint(0, 255)
        )

        if mac in macs:
            self.__change_mac(macs)
        return mac

    def __check_mac(self, mac):
        hwaddrs = []
        for cont, data in self.lxc_contrs.items():
            hwaddrs.append(data.get('hwaddr'))

        if mac in hwaddrs:
            mac = self.__change_mac(hwaddrs)
        return mac

    def _rw_container_run_sh(self):
        logging.debug('LXC cont name: {}'.format(self.p_unpck))

        for f_name in self.f_dwnld:
            try:
                if not '.tar.xz' == f_name[-7:]:
                    logging.info('Rewrite {} run file.'.format(f_name))
                    for target_name, cont_name in self.p_unpck.items():
                        if target_name in f_name:
                            conf_file = self.deff_lxc_cont_path + f_name
                            # Read run sh file
                            tmp_data = self.file_rw(
                                p=conf_file,
                                log='Read LXC {} run file'.format(f_name)
                            )
                            for row in tmp_data:
                                indx = tmp_data.index(row)
                                if 'CONTAINER_NAME=' in row:
                                    tmp_data[
                                        indx] = 'CONTAINER_NAME={}\n'.format(
                                        cont_name[0])
                                elif 'VPN_PORT=' in row:
                                    tmp_data[indx] = 'VPN_PORT={}\n'.format(
                                        self.use_ports['vpn'])

                            # rewrite run sh file
                            if not self.file_rw(
                                    p=conf_file,
                                    w=True,
                                    data=tmp_data,
                                    log='Rewrite LXC {} run file'.format(
                                        f_name)
                            ):
                                self._rolback(7)

                            del tmp_data
                            dest_run_sh_path = self.run_sh_path + f_name
                            logging.debug(
                                'LXC {} run file done'.format(f_name))
                            copyfile(conf_file, dest_run_sh_path)
                            logging.debug(
                                'LXC {} run file copy done'.format(f_name))
                            cmd = self.chmod_run_sh.format(dest_run_sh_path)
                            self._sys_call(cmd)
                            logging.debug(
                                'LXC {} run file chown done'.format(
                                    dest_run_sh_path))

            except BaseException as f_rw:
                logging.error('R/W LXC run sh: {}'.format(f_rw))
                self._rolback(32)

    @Init.wait_decor
    def _rw_container_intrfs(self):
        logging.debug('LXC containers: {}'.format(self.name_in_main_conf))
        for target_name, cont_name in self.p_unpck.items():
            try:
                conf_file = self.deff_lxc_cont_path + cont_name[
                    0] + self.lxc_cont_interfs
                # Read conf file
                tmp_data = self.file_rw(
                    p=conf_file,
                    log='Read LXC {} interfaces'.format(cont_name[0])
                )
                for row in tmp_data:
                    indx = tmp_data.index(row)

                    if 'address' in row:
                        tmp_data[indx] = 'address {}\n'.format(cont_name[1])
                    elif 'gateway' in row:
                        tmp_data[indx] = 'gateway {}\n'.format(
                            self.name_in_main_conf['LXC_ADDR='])
                    elif 'network' in row:
                        newrk = \
                            self.name_in_main_conf['LXC_NETWORK='].split(
                                '/')[0]
                        tmp_data[indx] = 'network {}\n'.format(newrk)

                # rewrite conf file
                if not self.file_rw(
                        p=conf_file,
                        w=True,
                        data=tmp_data,
                        log='Rewrite LXC {} interfaces'.format(cont_name)
                ):
                    self._rolback(7)

                del tmp_data

                logging.debug('LXC {} interfaces done'.format(cont_name))
                self._sys_call(self.update_cont_conf.format(conf_file))

            except BaseException as f_rw:
                logging.error('R/W LXC interfaces : {}'.format(f_rw))
                self._rolback(32)

    def _composit_addr(self, last_octet):
        addr = self.name_in_main_conf['LXC_ADDR='].split('.')
        addr[3] = last_octet
        addr = '.'.join(addr)
        return addr

    def _rw_psql_conf(self):
        logging.debug('Begin check DB configs')
        self.db_conf_path = self.db_conf_path.format(self.path_com)

        db_configs = {
            'pg_hba.conf':
                dict(
                    l_from=self.addr + "/24",
                    l_to=self.name_in_main_conf['LXC_NETWORK=']
                ),
            'postgresql.conf':
                dict(
                    l_from=self.def_comm_addr,
                    l_to=self.p_unpck['common'][1]
                )
        }
        logging.debug('Containers: {}'.format(self.p_unpck))
        logging.debug('Name in lxc conf: {}'.format(self.name_in_main_conf))
        logging.debug('DB configs: {}'.format(db_configs))
        for p_conf, fields in db_configs.items():
            p = self.db_conf_path + p_conf
            raw_data = self.file_rw(p=p, log='Read {} conf'.format(p_conf))
            for row in raw_data:
                if fields['l_from'] in row:
                    indx = raw_data.index(row)
                    raw_data[indx] = row.replace(fields['l_from'],
                                                 fields['l_to'])
            self.file_rw(p=p, log='Write db conf', w=True, data=raw_data)

    def _rw_container_conf(self):
        for target_name, cont_name in self.p_unpck.items():
            logging.debug('Rewrite {} conf'.format(cont_name[0]))
            try:
                conf_file = self.deff_lxc_cont_path + cont_name[
                    0] + self.lxc_cont_conf_name
                # Read conf file
                tmp_data = self.file_rw(
                    p=conf_file,
                    log='Read LXC {} config'.format(cont_name[0])
                )
                ipv4_indx = False
                for row in tmp_data:
                    indx = tmp_data.index(row)

                    if 'lxc.rootfs.path' in row:
                        tmp_data[
                            indx] = 'lxc.rootfs.path = dir:{}{}rootfs\n'.format(
                            self.deff_lxc_cont_path, cont_name[0])
                    elif 'lxc.uts.name' in row:
                        cont_name[0] = cont_name[0][0:-1]
                        tmp_data[indx] = 'lxc.uts.name = {}\n'.format(
                            cont_name[0])
                    elif 'lxc.net.0.hwaddr' in row:
                        hwaddr = row.split('=')[1].strip()
                        hwaddr = self.__check_mac(hwaddr)
                        tmp_data[indx] = 'lxc.net.0.hwaddr = {}\n'.format(
                            hwaddr)
                    elif 'lxc.net.0.ipv4.gateway' in row:
                        tmp_data[
                            indx] = 'lxc.net.0.ipv4.gateway = {}\n'.format(
                            self.name_in_main_conf['LXC_ADDR='])
                    elif 'lxc.net.0.ipv4.address' in row:
                        ipv4_indx = indx

                addr = self._composit_addr(str(cont_name[1]))
                self.p_unpck[target_name][1] = addr

                raw_ip_line = 'lxc.net.0.ipv4.address = {}/24\n'.format(addr)

                if ipv4_indx:
                    tmp_data[ipv4_indx] = raw_ip_line
                else:
                    tmp_data.append(raw_ip_line)

                # rewrite conf file
                if not self.file_rw(
                        p=conf_file,
                        w=True,
                        data=tmp_data,
                        log='Rewrite LXC {} config'.format(cont_name)
                ):
                    self._rolback(7)

                del tmp_data

                logging.debug('LXC {} config done'.format(cont_name))
                self._sys_call(self.update_cont_conf.format(conf_file))

            except BaseException as f_rw:
                logging.error('R/W LXC config : {}'.format(f_rw))
                self._rolback(32)

    def __check_contrs_by_cmd(self):
        logging.debug('Check containers by cmd')
        beg_line = 'NAME'
        pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$'

        raw = self._sys_call(cmd=self.exist_contrs, rolback=False)
        if raw:
            raw = raw.split('\n')
            marker = False
            for i in range(len(raw)):
                if raw[i].startswith(beg_line):
                    marker = i
                    break
            if len(raw) - 1 == marker:
                logging.debug('Containers are absent')
            else:
                for line in raw[marker + 1:]:
                    if line:
                        raw_line = line.split(' ')
                        for i in raw_line:
                            if i and i[-1] == ',':
                                i = i[:-1]
                            if i and match(pattern, i):
                                logging.debug(
                                    'Found new container: {}'.format(
                                        raw_line[0]))
                                if not self.lxc_contrs.get(raw_line[0]):
                                    self.lxc_contrs[raw_line[0]] = {
                                        'lxc.network.ipv4': i}
        else:
            logging.debug('Containers are absent')

    def _check_contrs_by_path(self, update_data=False):
        list_dir = listdir(self.deff_lxc_cont_path)
        folders = []
        for f in list_dir:
            if isdir(self.deff_lxc_cont_path + f):
                config_path = '{}{}'.format(
                    self.deff_lxc_cont_path,
                    f,
                )
                folders.append(config_path)
        # todo add folder where may be stored user containers
        # folders.append('')


        logging.debug('Check by path: {}'.format(folders))
        # read config file in container.Get data from it
        for folder in folders:
            tmp_data_name = None
            conf_path = '{}/{}'.format(folder, self.lxc_cont_conf_name)
            logging.debug('Conf in path: {}'.format(conf_path))
            if isfile(conf_path):
                res = self.file_rw(p=conf_path, log='Read container conf')
                tmp_data = {}
                for line in res:
                    for v in self.name_in_contnr_conf:
                        if v in line:
                            clear_line = line.split('=')[1].strip()
                            if v == 'lxc.network.ipv4':
                                clear_line = clear_line.split('/')[0]
                            tmp_data[v] = clear_line

                tmp_data_name = tmp_data['lxc.uts.name']
                del tmp_data['lxc.uts.name']
                if tmp_data_name in self.lxc_contrs:
                    self.lxc_contrs[tmp_data_name].update(tmp_data)
                else:
                    self.lxc_contrs[tmp_data_name] = tmp_data

            # exist file /rootfs/home/ubuntu/go/bin/dappctrl
            # search for the dappctrl file, by which we determine that
            # this is our container
            logging.debug(
                'Check container: {}'.format(folder + self.lxc_cont_fs_file))
            if isfile(folder + self.lxc_cont_fs_file) and tmp_data_name:
                logging.debug('Our container: {}'.format(tmp_data_name))
                # check what kind of container it is, vpn or common
                for c_name, c_path in self.kind_of_cont.items():
                    logging.debug('Check exist: {}'.format(folder + c_path))
                    if exists(folder + c_path):
                        logging.debug(
                            'Path: {} are exist'.format(folder + c_path))
                        if update_data:
                            self.p_unpck[c_name].append(
                                folder.split('/')[-1])
                            logging.debug(
                                'Update: {}'.format(self.p_unpck[c_name]))
                        else:
                            logging.info(
                                '\nWe found in your mashine installed our,\n '
                                '`Privatix` container: {} ,'
                                'was called in lxc as: {}.\nPlease remove it, '
                                'and repeat instalation.\n'
                                'Or re-run initializer in update mode!'.format(
                                    c_name, tmp_data_name)
                            )
                            sys.exit(31)
            else:
                logging.debug('Our container is absend')

    def __read_lxc_conf(self):
        # Check if exist /etc/default/lxc-net and 'LXC_BRIDGE' in it
        if isfile(self.bridge_conf):
            # conf exist, check it and search key from name_in_main_conf
            raw = self.file_rw(p=self.bridge_conf, log='Check bridge')
            if raw:
                for row in raw:
                    for k in self.name_in_main_conf:
                        if search('^' + k, row):
                            self.name_in_main_conf[k] = sub('"|{|}|\n', '',
                                                            row.split(k)[1])
                            logging.info(
                                'Found the bridge in the config: '
                                '{}{}'.format(k, self.name_in_main_conf[k]))

    def __check_lxc_exist(self):
        # Check if exist lxcbr bridge
        raw = self._sys_call(self.bridge_cmd)
        raw_arr = compile("\n\d:").split(raw)
        for row in raw_arr:
            if self.search_name in row:
                self.name_in_main_conf['LXC_BRIDGE='] = row.split(':', 1)[0]

        self.__read_lxc_conf()
        if self.name_in_main_conf['LXC_BRIDGE=']:
            logging.info('LXC already installed on computer')
            self._lxc_exist = True
            return True
        else:
            logging.info('LXC is not installed on computer.Installing it.')
            self._lxc_exist = False
            return False

    def __rename_cont_path(self):
        similar_contr = set(self.lxc_contrs.keys()) & set(
            self.kind_of_cont.keys())
        logging.debug('Rename path. Similar: {}'.format(similar_contr))
        for name in similar_contr:
            if name in self.path_vpn:
                self.path_vpn = 'dapp' + self.path_vpn
                self.p_unpck['vpn'][0] = self.path_vpn
            elif name in self.path_com:
                self.path_com = 'dapp' + self.path_com
                self.p_unpck['common'][0] = self.path_com

        self.db_log = self.db_log.format(self.path_com)
        logging.debug('Rename container path: {}'.format(self.p_unpck))

    def __check_wget(self):
        cmd = main_conf['search_pack'].format('wget')
        raw = self._sys_call(cmd=cmd)
        if not raw:
            logging.info('Install wget.')
            self._sys_call(cmd='sudo apt install wget')

    def _lxc_ubu_pack(self):
        self.__check_wget()
        if self.__check_lxc_exist():
            # lxc installed

            self.__check_contrs_by_cmd()
            self._check_contrs_by_path()
            logging.debug('Found LXC conteiners: {}'.format(self.lxc_contrs))
            self.__rename_cont_path()
            self._check_cont_addr()

        else:
            #     # lxc not installed
            self.__install_lxc()
            # update to new config params
            self.__read_lxc_conf()


def checker_fabric(inherit_class, old_vers, ver, dist_name):
    class Checker(Rdata, GUI, inherit_class):
        def __init__(self, old_vers, ver, dist_name):
            GUI.__init__(self)
            Rdata.__init__(self)
            inherit_class.__init__(self)
            self.old_vers = old_vers
            self.ver = ver
            self.dist_name = dist_name

        def __ubuntu(self):
            logging.debug('Ubuntu: {}'.format(self.ver))
            v = int(self.ver.split('.')[0])
            if v >= 16:
                logging.debug('--- Nspawn ---')
                self._nsp_ubu_pack()

            elif v >= 14:
                logging.debug('--- LXC ---')
                self._lxc_ubu_pack()

            else:
                logging.error('Your version of Ubuntu is not suitable. '
                              'It is not supported by the program')
                sys.exit(2)

        def __debian(self):
            logging.debug('Debian: {}'.format(self.ver))
            self._nsp_deb_pack()

        @Init.wait_decor
        def _check_os(self):
            logging.debug('Check OS')
            self.task = dict(ubuntu=self.__ubuntu,
                             debian=self.__debian
                             )
            try:
                task_os = self.task.get(self.dist_name.lower(), False)
                if not task_os:
                    mess = 'You system are {}.It is not supported yet'.format(
                        self.dist_name)
                    logging.error(mess)
                    return False, mess
                task_os()

                self.sysctl = self._sysctl() if not self.old_vers else True
                return True, ''
            except BaseException as checkExpt:
                logging.debug('Check OS: {}'.format(checkExpt))
                return False, checkExpt

    return Checker(old_vers, ver, dist_name)


def mainInitialCycle(log):
    global logging
    logging = log
    dist_name, ver, name_ver = linux_distribution()
    dist_name = dist_name.lower()
    old_vers = True if dist_name == 'ubuntu' and int(
        ver.split('.')[0]) < 16 else False
    if old_vers:
        check = checker_fabric(LXC, old_vers, ver, dist_name)
    else:
        check = checker_fabric(Nspawn, old_vers, ver, dist_name)
    return check
