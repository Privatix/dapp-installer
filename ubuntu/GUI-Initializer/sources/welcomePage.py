# -*- coding: utf-8 -*-


import logging
import requests
from json import load as jsonload
from time import sleep
from re import findall, search
from os.path import isfile, isdir
from os import remove, mkdir, system

from PyQt5 import QtCore, QtGui, QtWidgets
from PyQt5.QtCore import QThread, pyqtSignal
from PyQt5.QtWidgets import \
    QInputDialog, QWidget, QAction, QComboBox, \
    QLabel, QLineEdit, QMessageBox, QVBoxLayout, QTextEdit
from PyQt5.QtGui import QIcon, QFont
from waitingspinnerwidget import QtWaitingSpinner
from pexpect import spawn, EOF, TIMEOUT

from initializer import mainInitialCycle
from preparation import Preparation
import resource

logging.getLogger().setLevel('DEBUG')
form_console = logging.Formatter(
    '%(message)s',
    datefmt='%m/%d %H:%M:%S')

form_file = logging.Formatter(
    '%(levelname)7s [%(lineno)3s] %(message)s',
    datefmt='%m/%d %H:%M:%S')

fh = logging.FileHandler('initializer.log')  # file debug
fh.setLevel('DEBUG')
fh.setFormatter(form_file)
logging.getLogger().addHandler(fh)

ch = logging.StreamHandler()  # console debug
ch.setLevel('DEBUG')
ch.setFormatter(form_console)
logging.getLogger().addHandler(ch)


class ClassThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, meth, cmd=None):
        QThread.__init__(self)
        self.meth = meth
        self.cmd = cmd

    def run(self):
        if self.cmd and isinstance(self.cmd, list):
            res = self.meth(*self.cmd)
        elif self.cmd:
            res = self.meth(self.cmd)
        else:
            res = self.meth()

        if res:
            self.signal.emit((True, res))
        else:
            self.signal.emit((False, res))

#
#
# class PswdThread(QThread):
#     signal = pyqtSignal(tuple)
#
#     def __init__(self, meth, cmd=None):
#         QThread.__init__(self)
#         self.meth = meth
#         self.cmd = cmd
#
#     def run(self):
#         if self.cmd:
#             res = self.meth(self.cmd)
#         else:
#             res = self.meth()
#
#         if res:
#             self.signal.emit((True, res))
#         else:
#             self.signal.emit((False, res))
#

class CheckOsThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, meth):
        QThread.__init__(self)
        self.meth = meth

    def run(self):
        res = self.meth()
        self.signal.emit(res)

class DeleteThread(QThread):
    signal = pyqtSignal(str)

    def __init__(self, initial):
        QThread.__init__(self)
        self.initial = initial

    def run(self):
        mess = 'Stop and delete containers.'
        self.signal.emit(mess)
        self.initial.clear_contr(True)
        mess = 'Delete gui.'
        self.signal.emit(mess)
        self.initial._clear_dir(self.initial.gui_path)
        mess = 'Delete PID file.'
        self.signal.emit(mess)
        self.initial._clear_dir(self.initial.fin_file)
        mess = 'Delete Unit file.'
        self.signal.emit(mess)
        for unit in [self.initial.unit_f_com, self.initial.unit_f_vpn]:
            unit_dest = self.initial.unit_dest + unit
            mess = 'Delete {}.'.format(unit_dest)
            self.signal.emit(mess)
            self.initial._clear_dir(unit_dest)
        mess = 'Delete Icon file.'
        self.signal.emit(mess)
        self.initial._clear_dir(self.initial.gui_icon)
        self.signal.emit('0')


class ServicesThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, initial, change_perm):
        QThread.__init__(self)
        self.initial = initial
        self.change_perm = change_perm

    def run(self):
        try:
            self.signal.emit((True, 'Run common service.<br>Please wait.'))
            self.initial.run_service(comm=True)
            sleep(3)
            self.signal.emit((True, 'Check data base run.<br>Please wait.'))
            # system('sudo chmod 677 {}'.format(self.initial.db_log))
            res = self.initial._check_db_run(9)
            if res:
                self.signal.emit(
                    (True, 'Generate dappvpn.config.json.<br>Please wait.'))
                self.initial._run_dapp_cmd()
                self.signal.emit(
                    (True, 'Check dappvpn.config.json.<br>Please wait.'))
                self.change_perm(cycle='dapp')
                self.initial._check_dapp_conf()
                self.signal.emit(
                    (True, 'Reboot the common to apply the settings.<br>Please wait.'))

                if self.initial.dappctrl_role == 'agent':
                    # self.change_perm(cycle='tor_host')
                    self.initial.get_onion_key()
                else:
                    self.initial.set_socks_list()

                self.initial.run_service(comm=True, restart=True)

                self.signal.emit((True, 'Run vpn service.<br>Please wait.'))
                self.initial.run_service()
                self.signal.emit((True, 'Install GUI.<br>Please wait.'))
                res = self.initial.install_gui(self.change_perm)
                if res[0]:
                    self.signal.emit((True, 'end cycle', res[1]))
                else:
                    self.signal.emit((False, res[1]))


            else:
                self.signal.emit((False, 'Problem with DB'))
                self.spinner.stop()
        except BaseException as threxpt:
            self.signal.emit((False, threxpt))


class DownloadThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, fdwnld, url_dwnld, p_contr):
        QThread.__init__(self)
        self.f_dwnld = fdwnld
        self.url_dwnld = url_dwnld
        self.p_contr = p_contr

    def run(self):

        for f in self.f_dwnld:
            dwnld_url = self.url_dwnld + '/' + f
            dwnld_url = dwnld_url.replace('///', '/')

            with open(self.p_contr + f, "wb") as ftmp:
                mess = "Downloading {}.<br>Please wait.".format(f)
                response = requests.get(dwnld_url, stream=True)
                total_length = response.headers.get('content-length')
                if total_length is None:  # no content length header
                    ftmp.write(response.content)
                else:
                    dl = 0
                    total_length = int(total_length)
                    for data in response.iter_content(chunk_size=4096):
                        dl += len(data)
                        ftmp.write(data)
                        done = int(100 * dl / total_length)
                        self.signal.emit((done, mess))

        self.signal.emit((101, 'All download done'))


class WaitUpThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, initial):
        QThread.__init__(self)
        self.initial = initial

    def run(self):
        res = self.initial._wait_up()
        if res[0]:
            self.initial._finalizer(rw=True)
            self.signal.emit(('0', ''))
        else:
            self.signal.emit(('1', res[1]))


class RollbackThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, initial):
        logging.debug('RollbackThread init')
        QThread.__init__(self)
        self.initial = initial
        self.dumpPath = [
            '{}common'.format(self.initial.p_contr),
            '{}vpn/root/go/bin'.format(self.initial.p_contr),
            '{}vpn/opt/privatix/config'.format(self.initial.p_contr),
        ]

    def rolbackContainerData(self):
        logging.debug('Rollback containers data')
        for p in self.dumpPath:
            p_src = self.initial.contTmp + '/' + p.split('/')[-1]
            cmd = 'sudo cp -rf {} {}'.format(p_src, p)
            self.initial._sys_call(cmd, rolback=False)

        logging.debug('Try run containers after rollback')
        self.initial.run_service(comm=True)
        self.initial.run_service()

    def clearTmp(self):
        logging.debug('Clear tmp data')
        cmd = 'sudo rm -rf {}'.format(self.initial.contTmp)
        self.initial._sys_call(cmd=cmd, rolback=False)

    def run(self):
        logging.debug('Rollback run')
        self.signal.emit(('0', 'Preparation for stop all containers.Please wait.'))
        self.initial.stop_services()

        self.rolbackContainerData()
        self.clearTmp()
        self.signal.emit(('0', 'Data restored.'))


class UpdaterThread(RollbackThread):
    signal = pyqtSignal(tuple)

    def __init__(self, initial):
        RollbackThread.__init__(self, initial)
        logging.debug('UpdaterThread init')
        self.dwnldUpdateLink = 'http://art.privatix.net/binary/'
        self.dwnldFiles = {'dappctrl': ['common'],
                           'dappvpn': ['common', 'vpn']}

    def dumpContainerData(self):
        logging.debug('Prepare to dump data')
        cmd = 'sudo mkdir {0} && sudo chmod 777 {0}'.format(self.initial.contTmp)
        self.initial._sys_call(cmd, rolback=False)

        mess = 'Copying important files. It may take some time. Do not interrupt the process.'
        self.signal.emit(('0', mess))

        for p in self.dumpPath:
            cmd = 'sudo cp -rf {} {}'.format(p, self.initial.contTmp)
            logging.debug('Dump: {}'.format(cmd))
            if int(system(cmd)):
                logging.debug('Trouble when try dump data')
                return False, cmd
        return True, ''

    def downloadNewData(self):
        logging.debug('Download new data')
        for f in self.dwnldFiles:
            dwnld_url = self.dwnldUpdateLink + f
            self.signal.emit(('0', 'status_bar', 'start'))

            with open(self.initial.contTmp + '/' + f, "wb") as ftmp:
                mess = "Downloading {}.<br>Please wait.".format(f)
                self.signal.emit(('0', mess))
                response = requests.get(dwnld_url, stream=True)
                total_length = response.headers.get('content-length')
                if total_length is None:  # no content length header
                    ftmp.write(response.content)
                else:
                    dl = 0
                    total_length = int(total_length)
                    for data in response.iter_content(chunk_size=4096):
                        dl += len(data)
                        ftmp.write(data)
                        done = int(100 * dl / total_length)
                        self.signal.emit(('0', 'status_bar', done))

            self.signal.emit(('0', 'status_bar', 'stop'))

    def migrationNewData(self):
        logging.debug('Migrate container data')

        for f, contrs in self.dwnldFiles.items():
            for con in contrs:
                logging.debug('Migrate {} in {}'.format(f, con))
                p_dest = self.initial.p_contr + con + '/root/go/bin/'
                p_src = self.initial.contTmp + '/' + f
                cmd = 'sudo cp -f {} {}'.format(p_src, p_dest)
                if int(system(cmd)):
                    return False

        logging.debug('Migrate DB')
        # todo
        return True

    def updateNewData(self):
        logging.debug('Get new data')
        self.downloadNewData()
        return self.migrationNewData()

    def run(self):
        logging.debug('Updater run')
        self.signal.emit(('0', 'Preparation for stop all containers.Please wait.'))
        self.initial.stop_services()
        self.signal.emit(('0', 'Preparation for dumping information.Please wait.'))
        res = self.dumpContainerData()
        if not res[0]:
            self.signal.emit(('1', 'Trouble when try {}.<br>'
                                   'Further work cannot be continued.'.format(res[1])))

        else:
            self.signal.emit(('0', 'Preparation for download and update '
                                   'new data.<br>Please wait.'))

            if self.updateNewData():
                self.initial.run_service(comm=True)
                self.initial.run_service()
                self.signal.emit(('0', '0'))
            else:
                logging.debug('Trouble when try update data.Rollback.')
                self.rolbackContainerData()
                self.signal.emit(('1', 'An error has occurred.<br>'
                                       'Your data was not affected,'
                                       'it was saved and resumed.<br>'
                                       'Try again.'))


class UnpackThread(QThread):
    signal = pyqtSignal(tuple)

    def __init__(self, fdwnld, p_unpck, p_contr):
        QThread.__init__(self)
        self.f_dwnld = fdwnld
        self.p_unpck = p_unpck
        self.p_contr = p_contr

    def run(self):

        try:
            for f in self.f_dwnld:
                if '.tar.xz' == f[-7:]:
                    mess = 'Unpacking {}.<br>Please wait.'.format(f)
                    self.signal.emit((True, mess))

                    for k, v in self.p_unpck.items():
                        if k in f:
                            if not isdir(self.p_contr + v[0]):
                                mkdir(self.p_contr + v[0])
                            cmd = 'sudo tar xpf {} -C {} --numeric-owner'.format(
                                self.p_contr + f, self.p_contr + v[0])
                            system(cmd)

            self.signal.emit((True, 'end cycle'))

        except BaseException as expt_unpck:
            mess = 'Unpack: {}.'.format(expt_unpck)
            self.signal.emit((False, mess))


class InitGUI(QWidget):
    _translate = QtCore.QCoreApplication.translate
    _sysPswd = None

    def __init__(self, MainW,_updtRck):
        QWidget.__init__(self)
        self.mainPageUI(MainW)
        self.initial = mainInitialCycle(log=logging)
        self.rollbackEvent = None
        self._updateRollback = _updtRck  # declared in UpdateReinstall class


    def progrBar(self, on=False):
        if on:
            self.progressBar = QtWidgets.QProgressBar(self.centralwidget)
            self.progressBar.setGeometry(QtCore.QRect(240, 360, 321, 23))
            self.progressBar.setProperty("value", 0)
            self.progressBar.setObjectName("progressBar")
            self.progressBar.show()
            return self.progressBar.setProperty
        else:
            self.progressBar.close()

    def inputFld(self, on=False):
        if on:
            self.inputField = QtWidgets.QPushButton(self.centralwidget)
            self.inputField.setGeometry(QtCore.QRect(280, 220, 230, 40))
            self.inputField.setMaximumSize(QtCore.QSize(150, 16777215))
            self.inputField.setObjectName("inputField")
            self.inputField.setText("Press to input.")
            self.inputField.show()
        else:
            self.inputField.close()

    def UserEvent(self, event=None, typeEvent=None):
        logging.debug('UserEvent. Type: {}'.format(typeEvent))

        if typeEvent == 'Cancel':
            yesChoise = QtCore.QCoreApplication.instance().quit
            noChoise = lambda :'This is a blank stub'

        else:
            yesChoise = event.accept
            noChoise = event.ignore

        reply = QMessageBox.question(self,
                                     'Attention!',
                                     "Do you really want to stop the process and exit?",
                                     QMessageBox.Yes | QMessageBox.No,
                                     QMessageBox.No)
        if reply == QMessageBox.Yes:
            if self.rollbackEvent == 'new':
                self.initial._rolback(33)
            elif self.rollbackEvent == 'update':
                self._updateRollback()

            yesChoise()
        else:
            noChoise()


    def mainPageUI(self, MainWindow):
        MainWindow.setObjectName("MainWindow")
        MainWindow.resize(600, 490)
        MainWindow.setMinimumSize(QtCore.QSize(600, 520))
        MainWindow.setMaximumSize(QtCore.QSize(600, 520))
        MainWindow.setStyleSheet("")
        MainWindow.setWindowTitle("Privatix")
        MainWindow.setWindowIcon(QIcon(":/images/icon/icon.png"))

        self.centralwidget = QtWidgets.QWidget(MainWindow)
        self.centralwidget.setObjectName("centralwidget")
        self.pageText = QtWidgets.QTextBrowser(self.centralwidget)
        self.pageText.setGeometry(QtCore.QRect(225, 10, 350, 400))
        self.pageText.setMaximumSize(QtCore.QSize(16777215, 400))
        self.pageText.setObjectName("pageText")
        self.pageText.setHtml(self._translate("MainWindow",
                                              "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.0//EN\" \"http://www.w3.org/TR/REC-html40/strict.dtd\">\n"
                                              "<html><head><meta name=\"qrichtext\" content=\"1\" /><style type=\"text/css\">\n"
                                              "p, li { white-space: pre-wrap; }\n"
                                              "</style></head><body style=\" font-family:\'Sans Serif\'; font-size:9pt; font-weight:400; font-style:normal;\">\n"
                                              "<p style=\" margin-top:0px; margin-bottom:0px; margin-left:0px; margin-right:0px; -qt-block-indent:0; text-indent:0px;\">"
                                              "<span style=\" font-size:12pt; font-style:italic;\">"
                                              "    Welcome.<br>For further installation it will "
                                              "be necessary to enter your root password.<br>"
                                              "    Also, to work with the GUI version of the program, "
                                              "need to install the nodejs versions no lower than 9.0 and npm not lower than 5.6, "
                                              "they will be installed automatically.<br>"
                                              "    If you already have these packages installed "
                                              "with versions that do not satisfy the requirement,"
                                              "you need to click `Cancel` and remove them manually. "
                                              "The next start the application, "
                                              "they will be automatically installed on your machine.<br>"
                                              "    Click `Next` if you agree and the requirements are met.</span></p></body></html>"))

        self.pageNext = QtWidgets.QPushButton(self.centralwidget)
        self.pageNext.setGeometry(QtCore.QRect(480, 420, 91, 41))
        self.pageNext.setMaximumSize(QtCore.QSize(100, 16777215))
        self.pageNext.setObjectName("pageNext")
        self.pageNext.setText(self._translate("MainWindow", "Next"))

        self.pageQuit = QtWidgets.QPushButton(self.centralwidget)
        self.pageQuit.setGeometry(QtCore.QRect(370, 420, 91, 41))
        self.pageQuit.setMaximumSize(QtCore.QSize(100, 16777215))
        self.pageQuit.setObjectName("pageQuit")
        self.pageQuit.clicked.connect(lambda: self.UserEvent(typeEvent='Cancel'))
        self.pageQuit.setText(self._translate("MainWindow", "Cancel"))

        self.welcomeLogo = QtWidgets.QLabel(self.centralwidget)
        self.welcomeLogo.setGeometry(QtCore.QRect(20, 10, 200, 400))
        self.welcomeLogo.setPixmap(QtGui.QPixmap(":/images/icon/main.png"))
        self.welcomeLogo.setObjectName("welcomeLogo")

        MainWindow.setCentralWidget(self.centralwidget)
        self.statusbar = QtWidgets.QStatusBar(MainWindow)
        self.statusbar.setObjectName("statusbar")
        MainWindow.setStatusBar(self.statusbar)

        QtCore.QMetaObject.connectSlotsByName(MainWindow)

        self.spinner = QtWaitingSpinner(self)
        self.statusbar.addWidget(self.spinner)

        quit = QAction("Quit", self)
        quit.triggered.connect(self.close)
        # self.tray_icon = QSystemTrayIcon(self)
        # self.tray_icon.setIcon(self.style().standardIcon(QStyle.SP_ComputerIcon))
        # self.tray_icon.show()
        self.initInteractiveLayout()

    def initInteractiveLayout(self):
        # layout = QVBoxLayout(self.centralwidget)
        self.interLayout = QTextEdit(self.centralwidget)
        self.interLayout.setGeometry(225, 10, 350, 400)
        # layout.addWidget(self.interLayout)
        # self.setLayout(layout)
        self.interLayout.hide()

    def showStreamInterLayout(self, proc):
        mess = str(proc.readAllStandardOutput())
        self.interLayout.append(mess)
        print mess


class UpdateReinstall(InitGUI):
    def __init__(self, MainW):
        InitGUI.__init__(self, MainW, self._updateRollback)

    def finish(self, upt=False):
        """ Abstraction. Redefined in Prepare class """
        pass

    def updateReinstall(self):
        """ Update, reinstall, delete."""
        self.job = None
        self.__choisePage()

    def __choisePage(self):
        logging.debug('Show __choisePage on second start')

        def prepUpdate():
            self.rollbackEvent = 'update'
            mess = "You have chosen to update the software.<br>" \
                   "Make sure that you have saved all the data,<br>" \
                   "otherwise they will be lost forever!"
            self.job = self.__updateJob, 'update'
            return mess

        def prepReinstall():
            self.rollbackEvent = 'reinstall'
            mess = "You have chosen to completely reinstall<br>the software," \
                   "during which all data will be deleted<br>and reinstalled.<br>" \
                   "Make sure you save all the data,<br>" \
                   "otherwise they will be lost forever!"
            self.job = self.__reinstallJob, 'reinstall'
            return mess

        def prepDelete():
            self.rollbackEvent = 'delete'
            mess = "You have chosen to completely delete<br>the software," \
                   "during which all data will be deleted<br>from your machine.<br>" \
                   "Make sure you save all the data,<br>" \
                   "otherwise they will be lost forever!"
            self.job = self.__deleteJob, 'delete'
            return mess

        def select_flow(text):
            self.show_mess.setText(task[text]())
            self.show_mess.show()
            self.pageNext.show()

        def __doJob():
            logging.debug('Chaise: {}'.format(self.job[1]))
            if self.job[1] == 'delete':
                reply = QMessageBox.question(self,
                                             'Message',
                                             "Do you really want to delete the program?<br>"
                                             "If you have not saved the data,<br>"
                                             "they will be lost forever.",
                                             QMessageBox.Yes | QMessageBox.No,
                                             QMessageBox.No)
                if reply == QMessageBox.No:
                    logging.debug('No choise')
                    QtCore.QCoreApplication.instance().quit()

            self.pageComboBox.close()
            self.show_mess.close()
            self.pageNext.hide()
            self.pageText.setHtml(
                'Please wait, {} is in progress.'.format(self.job[1]))
            self.spinner.start()
            logging.debug('Prepare to run job: {}'.format(self.job[1]))
            self.job[0]()

        self.show_mess = QLabel(self.pageText)
        self.show_mess.setGeometry(QtCore.QRect(5, 150, 350, 100))

        task = dict(
            Update=prepUpdate,
            Reinstall=prepReinstall,
            Delete=prepDelete)

        self.pageText.setHtml('This is a restart of the initializers.<br>'
                              'Please choose what you want to update and click `Next`.')

        self.pageComboBox = QComboBox(self.pageText)
        self.pageComboBox.setGeometry(QtCore.QRect(10, 80, 150, 30))

        for t in task:
            self.pageComboBox.addItem(t)

        self.pageComboBox.activated[str].connect(select_flow)
        self.pageComboBox.show()
        self.pageNext.clicked.connect(__doJob)

    def _updateRollback(self):
        logging.debug('Rollback when update')

        def rollbackDone(resp):
            self.pageText.setHtml(resp[1])
            self.spinner.stop()

        self.thr = RollbackThread(initial=self.initial)
        self.thr.signal.connect(rollbackDone)
        self.thr.start()

    def __updateJob(self):
        logging.debug('Update Job')

        def updateJobDone(resp):
            if int(resp[0]):
                # trouble in cycle
                logging.debug('The process ended with an error: {}'.format(resp[1]))
                self.pageText.setHtml(
                    'An error occurred during operation.<br>{}'.format(resp[1]))
                self.spinner.stop()
                self.pageQuit.hide()
                self.pageNext.setText(self._translate("MainWindow", "Finish"))
                self.pageNext.clicked.connect(QtCore.QCoreApplication.instance().quit)
            else:
                if resp[1] == '0':
                    # cycle completed
                    logging.debug('Main cycle successfully completed.')
                    self.pageText.setHtml('Update data is complete.<br>'
                                          'Preparation for the launch of containers and check their work.<br>'
                                          'Wait for the process to complete.')

                    self.finish(upt=True)
                elif resp[1] == 'status_bar':
                    # draw status bar
                    if resp[2] == 'start':
                        logging.debug('progrBar start')
                        self.bar_act = self.progrBar(on=True)
                    elif resp[2] == 'stop':
                        logging.debug('progrBar stop')
                        self.progrBar()
                        del self.bar_act
                    else:
                        self.bar_act("value", resp[2])
                else:
                    # show user mess
                    self.pageText.setHtml(resp[1])

        self.spinner.start()
        self.thr = UpdaterThread(initial=self.initial)
        self.thr.signal.connect(updateJobDone)
        self.thr.start()

    def __reinstallJob(self):
        logging.debug('Reinstall Job')
        def reinstallJobDone(resp):
            logging.debug('All Deleted.')
            self.spinner.stop()
            self.pageNext.disconnect()
            self.startCycle(purge=True)

        def __clearDirs():
            logging.debug('Clear cont dirs')
            self.initial.clear_contr(True)
            logging.debug('Clear gui dirs')
            self.initial._clear_dir(self.initial.gui_path)
            self.initial.use_ports = dict(vpn=[], common=[],
                                          mangmt=dict(vpn=None,
                                                      common=None))
            logging.debug('Clear PID file')
            self.initial._clear_dir(self.initial.fin_file)

            return True

        self.thr = ClassThread(meth=__clearDirs)
        self.thr.signal.connect(reinstallJobDone)
        self.thr.start()

    def __deleteJob(self):
        logging.debug('Delete Job')
        def deleteJobDone(resp):
            if resp == '0':
                logging.debug('All Deleted.')
                self.spinner.stop()
                self.pageNext.disconnect()
                self.pageText.setHtml('All data has been completely deleted.')
                self.pageQuit.hide()
                self.pageNext.show()
                self.pageNext.setText(self._translate("MainWindow", "Finish"))
                self.pageNext.clicked.connect(QtCore.QCoreApplication.instance().quit)

            else:
                logging.debug(resp)
                self.pageText.setHtml(resp)


        self.thr = DeleteThread(initial=self.initial)
        self.thr.signal.connect(deleteJobDone)
        self.thr.start()


class Prepare(UpdateReinstall):
    def __init__(self, MainW):
        UpdateReinstall.__init__(self, MainW)
        self.pageNext.clicked.connect(self.check_install_pack)

    def check_install_pack(self):
        logging.debug('Check packs')
        self.pageNext.hide()

        res = self.initial.search_gui_pack()
        logging.debug('Gui: {}'.format(res))
        if res[0]:
            mess = 'You have installed obsolete packages.<br>' \
                   'For further work you need to delete them.' \
                   'And re-run the application<br>'
            for pack_mess in res[1]:
                mess = mess + pack_mess

            self.pageText.setHtml(mess)
            self.pageQuit.setText(self._translate("MainWindow", "Ok"))

        else:
            self.pageNext.disconnect()
            self.startCycle()

    def checkInst(self):
        logging.debug('Check inst')

        if isfile(self.initial.fin_file):
            try:
                file_data = open(self.initial.fin_file)
                raw = jsonload(file_data)
                self.initial.use_ports.update(raw)
                logging.debug('Read pid file done: {}'.format(raw))
                return True
            except BaseException as rwex:
                logging.debug('Trouble when read pid file: {}'.format(rwex))
                return True
        logging.debug('Clean installing')
        return False


    def startCycle(self, purge=False):

        # def install_done(res):
        #     self.spinner.stop()
        #
        #     if res[0]:
        #         self.pageText.setHtml(
        #             'The package installation was successful done.<br>'
        #             'Click `Next` to start the configuration.')
        #         self.pageNext.show()
        #         self.pageNext.clicked.connect(self.check_role)
        #
        #     else:
        #         self.pageText.setHtml('Sorry that something went wrong.<br>'
        #                               'The installation was canceled')

        def install_done(objResp):
            self.spinner.stop()
            exitCode = objResp.exitCode()

            logging.debug('Install done: {}'.format(exitCode))
            if not exitCode:
                self.pageText.setHtml(
                    'The package installation was successful done.<br>'
                    'Click `Next` to start the configuration.')
                self.pageNext.show()
                self.pageNext.clicked.connect(self.check_role)

            else:
                self.pageText.setHtml('Sorry that something went wrong.<br>'
                                      'The installation was canceled')

        if self.checkInst():
            logging.debug(' - Update reinstall delete')
            self.updateReinstall()
        else:
            logging.debug(' - New Install')
            self.rollbackEvent = 'new'
            chPsdw = self.enterPass()
            if chPsdw[0]:
                prepare = Preparation(
                    log=logging,
                    bar=self.progrBar,
                    page_t=self.pageText,
                    sys_pswd=chPsdw[1],
                    sys_call_meth=self.sysCallWithPswd
                )
                if purge:
                    self.initial._sys_call(cmd=prepare.del_pack)
                    # self.initial._clear_dir()
                chPrep = prepare.preparation()
                if chPrep[0]:
                    fin_res = self.initial._finalizer()
                    if fin_res[0]:
                        logging.debug('Install pack cmd: {}'.format(chPrep[1]))
                        self.spinner.start()
                        # self.thr = ClassThread(meth=self.initial._sys_call,
                        #                        cmd=chPrep[1])
                        # self.thr.signal.connect(install_done)
                        # self.thr.start()
                        self.interLayout.show()
                        self.process = QtCore.QProcess(self)
                        self.process.setProcessChannelMode(
                            self.process.MergedChannels)
                        self.process.readyReadStandardOutput.connect(
                            lambda: self.showStreamInterLayout(self.process))
                        self.process.start(chPrep[1])
                        self.process.finished.connect(lambda: install_done(self.process))

                    else:
                        logging.debug('Finalizer: {}'.format(fin_res))
                        self.pageText.setHtml(fin_res[1])

                else:
                    logging.error('Result after preparation:{}'.format(chPrep))
                    self.pageText.setHtml(str(chPrep[1]))
                    self.pageNext.close()
            else:
                self.startCycle()

    def sysCallWithPswd(self, cmd, pswd, only_check=None, manager=list()):
        logging.debug('Check pass cmd:{} Pswd: {}'.format(cmd, pswd))
        # todo - delete this return!
        # return True
        try:
            # session = spawn(command='ls')
            session = spawn(cmd, timeout=3)

        except BaseException as sespawn:
            logging.error('Spawn Error: {}'.format(sespawn))
            manager.append(False)
            return False

        logging.debug('Send pswd: {}'.format(pswd))
        if not only_check:
            session.sendline(pswd)
        try:
            session.expect([EOF], timeout=2)
        except TIMEOUT as sesexpt:
            logging.error('Timeout Error: {}'.format(sesexpt))
        logging.debug('session.before: {}'.format(session.before))
        if search('([0-9]{4})', session.before):
            logging.debug('Correct pswd')
            manager.append(True)
            return True

        logging.debug('Wrong pswd')
        manager.append(False)
        return False

    def pswdOnMultiProcc(self, cmd, pswd):
        logging.debug('pswdOnMultiProcc: {}'.format(cmd))
        from multiprocessing import Process, Manager
        manager = Manager()
        resp = manager.list()
        p = Process(target=self.sysCallWithPswd, args=(cmd, pswd, None, resp))
        p.start()
        # sleep(2)
        # p.terminate()
        p.join()
        print resp
        exit(888)

    def pswdOnThr(self, cmd, pswd):
        logging.debug('pswdThr: {}'.format(cmd))
        import threading

        resp = list()
        t1 = threading.Thread(target=self.sysCallWithPswd, args=(cmd, pswd))
        t1.start()
        t1.join()
        print resp
        exit(888)

    def enterPass(self):

        def checkPswdDone(res):
            logging.debug('Result sys call: {}'.format(res))
            if res:
                logging.debug('Pass was accept')
                self.pageText.setText(
                    'Password is correct.<br>Click `Next` please.')
                return True, sysPswd
            else:
                logging.debug('Pass was wrong')
                self.pageText.setText(
                    'Password is incorrect!<br>'
                    'Please try again.')
                return False, ''

        self.pageNext.close()
        self.pageText.setHtml('Enter your user root password!')

        sysPswd, ok = QInputDialog.getText(self,
                                           'Password',
                                           'Please enter your user root password:',
                                           QLineEdit.Password)

        if ok:
            if self.initial.dist_name == 'ubuntu':
                check_pass_cmd = 'sudo sh -c "date \'+%Y\'"'
            else:
                check_pass_cmd = 'su - root -c \'date \'+%Y\'\''
            logging.debug('root* pswd: {}'.format(sysPswd))
            res = self.sysCallWithPswd(cmd=check_pass_cmd, pswd=sysPswd)
            # res = self.pswdOnMultiProcc(check_pass_cmd,sysPswd)
            # res = self.pswdOnThr(check_pass_cmd,sysPswd)

            # self.thr = ClassThread(meth=self.sysCallWithPswd,cmd=[check_pass_cmd,sysPswd])
            # self.thr.signal.connect(checkPswdDone)
            # self.thr.start()
            # self.thr.wait()
            logging.debug('Result sys call: {}'.format(res))
            if res:
                logging.debug('Pass was accept')
                self.pageText.setText(
                    'Password is correct.<br>Click `Next` please.')
                return True, sysPswd
            else:
                logging.debug('Pass was wrong')
                self.pageText.setText(
                    'Password is incorrect!<br>'
                    'Please try again.')
                return False, ''

        else:
            logging.debug('Press quit on dialog input field')
            QtCore.QCoreApplication.instance().quit()
            return False, ''

    def change_perm(self, cycle=None):
        logging.debug('Change permissions')

        def perm(perm_files):
            for path in perm_files:
                cmd = "sudo stat -c '%a %n' {}".format(path)
                res = self.initial._sys_call(cmd=cmd)
                if res:
                    perm_code = res.split(' ')[0]
                    self.perm_files[path] = perm_code
                    cmd = 'sudo chmod 777 {}'.format(path)
                    self.initial._sys_call(cmd=cmd)
                else:
                    logging.error(' When try {} :{}'.format(cmd,res))

        p_cont = self.initial.p_contr

        path_com_conf = p_cont + self.initial.path_com + self.initial.p_dapctrl_conf
        path_dapvpn_conf = p_cont + self.initial.path_vpn + self.initial.p_dapvpn_conf
        path_dapcom_conf = p_cont + self.initial.path_com + self.initial.p_dapvpn_conf
        path_vpn_conf = p_cont + self.initial.path_vpn + self.initial.ovpn_conf
        path_vpn_unit = p_cont + self.initial.unit_f_vpn
        path_tor_h_conf = p_cont + self.initial.path_com + self.initial.tor_hostname_config
        path_tor_conf = p_cont + self.initial.path_com + self.initial.tor_config

        if cycle == 'dapp':
            perm_files = {
                path_dapvpn_conf: '',
                path_dapcom_conf: '',
            }
            perm(perm_files)
        elif cycle == 'config':
            perm_files = {
                path_vpn_conf: '',
                path_vpn_unit: '',
                path_com_conf: ''
            }
            perm(perm_files)

        elif cycle == 'gui':
            perm_files = {
                self.initial.dappctrlgui: ''
            }
            perm(perm_files)

        elif cycle == 'tor_host':
            perm_files = {
                path_tor_h_conf: '',
            }
            perm(perm_files)

        elif cycle == 'tor_conf':
            perm_files = {
                path_tor_conf: ''
            }
            perm(perm_files)
        else:
            for path, code in self.perm_files.items():
                cmd = 'sudo chmod {} {}'.format(code, path)
                self.initial._sys_call(cmd=cmd)
                del self.perm_files[path]

    def finish(self, upt=False):
        ''' upt for rolback when update or reinstall'''

        def finish_done(resp):
            self.pageNext.disconnect()
            self.spinner.stop()

            if resp[0] == '0':
                if upt:
                    self.initial._sys_call(
                        'sudo rm -rf {}'.format(self.initial.contTmp))

                self.pageNext.setEnabled(True)
                self.pageNext.show()
                self.pageText.setHtml(
                    'Application was start,and ready to go.<br>'
                    'Please click to `Finish`.<br>'
                    'Enjoy :)')
                self.pageNext.setText(self._translate("MainWindow", "Finish"))
                self.pageNext.clicked.connect(QtCore.QCoreApplication.instance().quit)
            else:
                if upt:
                    self.pageText.setHtml('Sorry, an error occurred during the work.<br>'
                                          'We saved your data and return everything to its original position.<br>'
                                          'Please wait for the operation to complete.')

                    self._updateRollback()
                else:
                    self.pageText.setHtml(resp[1])

        self.pageNext.setEnabled(False)
        if not upt:
            self.change_perm()
        self.spinner.start()
        self.pageText.setHtml(
            'Wait for the application to start.<br>'
            'This may take some time.')

        self.thr = WaitUpThread(initial=self.initial)
        self.thr.signal.connect(finish_done)
        self.thr.start()

    def main_cycle(self):
        logging.debug(' - main_cycle')

        def serv_done(resp):
            if resp[0]:
                if resp[1] == 'end cycle':
                    self.pageText.setHtml(
                        'The installation is complete.<br>'
                        '{}<br>'
                        'Click `Next` and wait '
                        'for the application to start.'.format(resp[2]))
                    self.spinner.stop()

                    self.pageNext.clicked.connect(self.finish)
                    self.pageNext.setEnabled(True)
                else:
                    self.pageText.setHtml(str(resp[1]))
            else:
                self.spinner.stop()
                self.pageText.setHtml(str(resp[1]))


        mess = 'Prepare configs.<br>Please wait, this may take a few minutes.'
        self.pageText.setHtml(mess)
        logging.debug(mess)
        self.perm_files = dict()
        self.change_perm(cycle='config')

        self.initial._rw_openvpn_conf(
            self.initial.addr,
            self.tun,
            self.initial.use_ports['vpn'],
            7)

        self.initial._rw_unit_file(self.initial.addr, self.intrfs, 5)

        self.initial.clean()
        self.initial._clear_db_log()

        self.initial.conf_dappctrl_json()
        self.change_perm(cycle='tor_conf')
        self.initial.check_tor_port()

        self.thr = ServicesThread(initial=self.initial,
                                 change_perm=self.change_perm)
        self.thr.signal.connect(serv_done)
        self.thr.start()


class Rdata(Prepare):
    def __init__(self, MainW):
        Prepare.__init__(self, MainW)

    def download(self):
        logging.debug(' - download')
        bar_act = self.progrBar(on=True)
        self.pageNext.disconnect()

        def dw_done(result):
            if result[0] > 100:
                self.progrBar()
                sleep(2)
                self.unpacking()
            else:
                bar_act("value", result[0])
            self.pageText.setHtml(result[1])

        self.pageNext.setEnabled(False)
        self.pageText.setHtml('Begin download files.<br>Please wait.')

        if not isdir(self.initial.p_contr):
            logging.debug('Create dir: {}'.format(self.initial.p_contr))
        system('sudo mkdir {0};sudo chmod 777 {0};'.format(
            self.initial.p_contr))

        self.thr = DownloadThread(self.initial.f_dwnld,
                                  self.initial.url_dwnld,
                                  self.initial.p_contr)
        self.thr.signal.connect(dw_done)
        self.thr.start()

    def unpacking(self):
        logging.debug(' - unpacking')

        def un_done(result):
            if result[0] and result[1] == 'end cycle':
                # self.spinner.stop()
                # system('sudo chmod -R 777 {0};'.format(
                #     self.initial.p_contr))
                try:
                    self.main_cycle()
                except BaseException as mainexpt:
                    self.pageText.setHtml('Oops Trouble.')
                    self.spinner.stop()
                    logging.error('Main cycle: {}'.format(mainexpt))


            else:
                self.pageText.setHtml(result[1])
                logging.info('Unpack done: {}'.format(result))

        mess = 'Begin unpacking download files.'
        logging.info(mess)
        self.pageText.setHtml(mess)
        self.spinner.start()

        self.thr = UnpackThread(self.initial.f_dwnld, self.initial.p_unpck,
                                self.initial.p_contr)
        self.thr.signal.connect(un_done)
        self.thr.start()

    def clean(self):
        logging.info('Delete downloaded files.')

        for f in self.f_dwnld:
            logging.info('Delete {}'.format(f))
            remove(self.p_contr + f)


class Interfaces(Rdata):
    def __init__(self, MainW,check_role):
        Rdata.__init__(self, MainW)
        self.check_role = check_role

    def addrCheck(self):
        def compare_addr(all_addr, pattern, new_addr):
            if not search(pattern, new_addr):
                mess = 'You addres {} is wrong,please enter ' \
                       'right address.Last octet is always 0.Example: 255.255.255.0\n'.format(
                    new_addr)
                logging.info(mess)
                return False, mess

            for i in all_addr:
                if new_addr + self.initial.mask[0] in i:
                    mess = 'Addres {} is busy or wrong, please enter new address ' \
                           'without changing the 4th octet.' \
                           'Example: xxx.xxx.xxx.0\n'.format(new_addr)
                    logging.info(mess)
                    return False, mess
            return True, new_addr

        self.pageText.setHtml('Check Addres: {}'.format(self.initial.addr))
        logging.debug('Check iptables')

        cmd = 'sudo /sbin/iptables -t nat -L'
        chain = 'Chain POSTROUTING'
        raw = self.initial._sys_call(cmd)
        arr = raw.split('\n\n')
        chain_arr = []
        for i in arr:
            if chain in i:
                chain_arr = i.split('\n')
                break
        del arr
        self.pageNext.setEnabled(False)
        for i in chain_arr:
            if self.initial.addr + self.initial.mask[0] in i:
                mess = 'Address {} is busy or wrong.<br>' \
                       'Please enter new address ' \
                       'without changing the 4th octet.<br>' \
                       'Example: xxx.xxx.xxx.0\n'.format(self.initial.addr)
                logging.info(mess)

                pattern = r'^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}0$'

                self.pageText.setHtml(mess)
                new_addr, ok = QInputDialog.getText(self, 'Enter new addr',
                                                    'Enter the correct address:')
                if ok:
                    self.pageNext.setEnabled(True)

                    result = compare_addr(chain_arr, pattern, str(new_addr))
                    if not result[0]:
                        logging.debug(
                            'Addr:{},Res:{}'.format(new_addr, result))
                        self.pageText.setHtml(result[1])
                    else:

                        self.initial.addr = result[1]
                        self.pageNext.setEnabled(True)
                        self.pageText.setHtml('Addres {} is free.<br>'
                                              'Please click `Next` '
                                              'for check your network type.'.format(
                            self.initial.addr))
                        logging.debug(
                            'Correct addr: {}'.format(self.initial.addr))
                        self.pageNext.disconnect()
                        self.pageNext.clicked.connect(self.netwCheck)


                    break
                else:
                    logging.debug('Choise NO')
                    QtCore.QCoreApplication.instance().quit()
        else:
            self.pageNext.setEnabled(True)
            self.pageText.setHtml('Addres {} is free.<br>'
                                  'Please click `Next` '
                                  'for check your network type'.format(
                self.initial.addr))
            self.pageNext.disconnect()
            self.pageNext.clicked.connect(self.netwCheck)

    def netwCheck(self):
        self.pageText.setHtml('Prepere to check network.')
        self.pageNext.setEnabled(False)

        self.tunCheck()

        def netw_check(text):
            self.intrfs = text

        arr_intrfs = []
        cmd = 'sudo LANG=POSIX lshw -C network'
        raw = self.initial._sys_call(cmd)
        arr = raw.split('logical name: ')
        arr.pop(0)
        for i in arr:
            arr_intrfs.append(i.split('\n')[0])
        del arr
        self.pageComboBox = QComboBox(self.centralwidget)
        self.pageComboBox.setGeometry(QtCore.QRect(350, 200, 100, 45))
        if len(arr_intrfs) > 1:
            self.intrfs = arr_intrfs[0]
            for intrf in arr_intrfs:
                self.pageComboBox.addItem(intrf)

            self.pageText.setHtml(
                'You have many interfaces,choose one of them and<br>'
                'click `Next` for continue installation.')
        else:
            self.intrfs = arr_intrfs[0]
            self.pageComboBox.addItem(self.intrfs)

            self.pageText.setHtml('You have one interface: {}.<br>'
                                  'Click `Next` for continue installation.'.format(
                self.intrfs))
        self.pageComboBox.activated[str].connect(netw_check)
        self.pageComboBox.show()

        self.pageNext.setEnabled(True)
        self.pageNext.disconnect()
        self.pageNext.clicked.connect(self.portCheck)

    def tunCheck(self):
        self.pageText.setHtml('Begin check tun interface in your system.')
        logging.debug("Tun check")

        def check_tun(i):
            max_tun_index = max([int(x.replace('tun', '')) for x in i])

            logging.info('You have the following interfaces {}.<br>'
                         'Please enter another tun interface.<br>'
                         'For example tun{}.\n'.format(i, max_tun_index + 1))

            return max_tun_index + 1

        cmd = 'sudo ip link show'
        raw = self.initial._sys_call(cmd)
        tuns = findall("tun\d", raw)
        tun = 'tun1'
        if tuns:
            tun = check_tun(tuns)
        self.tun = tun

    def portCheck(self):
        self.pageComboBox.hide()
        self.pageNext.setEnabled(False)

        logging.debug("Port check")
        self.pageText.setHtml('Begin check free port in your system.')

        port = self.initial.ovpn_port[0]
        port = findall('\d+', port)[0]

        port = self.initial.check_port(port=port, auto=True)

        self.initial.use_ports['vpn'] = port
        self.pageText.setHtml('Port {} is free.<br>We use it for VPN.'.format(port))
        self.pageNext.setEnabled(True)
        self.pageNext.disconnect()
        self.pageNext.clicked.connect(self.download)


class Configure(Interfaces):
    def __init__(self, MainW):
        Interfaces.__init__(self, MainW, self.check_role)

    def check_role(self):
        def select_role(text):
            self.initial.dappctrl_role = text.lower()
            logging.debug('Choosen role: {}'.format(self.initial.dappctrl_role))

        self.interLayout.hide()
        self.interLayout.destroy()
        self.initial.target = 'both'  # install back & gui
        # self.pageNext.setEnabled(False)

        self.pageText.setHtml('Begin configure.<br>'
                              'Please choose your role, and click `Next`.')

        self.pageComboBox = QComboBox(self.centralwidget)
        self.pageComboBox.setGeometry(QtCore.QRect(350, 200, 100, 45))

        self.initial.dappctrl_role = "agent"
        self.pageComboBox.addItem("Agent")
        self.pageComboBox.addItem("Client")
        self.pageComboBox.activated[str].connect(select_role)
        self.pageComboBox.show()
        self.pageNext.disconnect()
        self.pageNext.clicked.connect(self.check_os)

    def check_os(self):
        def check_os_done(res):
            self.spinner.stop()
            if res[0]:
                self.pageText.setHtml(
                    'All dependencies are installed, click `Next` to continue.')
                self.pageNext.show()
                self.pageNext.disconnect()
                self.pageNext.clicked.connect(self.addrCheck)
            else:
                self.pageText.setHtml(res[1])

        self.pageComboBox.hide()
        self.pageText.setHtml(
            'Checking your OS and install all dependencies.<br>Please wait.')

        self.pageNext.hide()

        self.spinner.start()
        self.thr = CheckOsThread(meth=self.initial._check_os)
        self.thr.signal.connect(check_os_done)
        self.thr.start()


if __name__ == "__main__":
    import sys

    app = QtWidgets.QApplication(sys.argv)
    MainWindow = QtWidgets.QMainWindow()
    ui = Configure(MainWindow)
    MainWindow.closeEvent = ui.UserEvent

    MainWindow.show()
    sys.exit(app.exec_())
