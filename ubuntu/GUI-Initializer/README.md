The following packages must be installed before compiling

    sudo apt install python2.7
    sudo apt install python2.7-dev -y
    sudo apt install python-pyqt5 -y
    sudo apt install pyqt5-dev-tools -y

    sudo pip2 install requests
    sudo pip2 install pexpect
    sudo pip2 install pyinstaller

To compile, go to the directory where the main Initializer.py is located,
and execute:
pyinstaller -F -w -i=/sources/icon/icon.png --clean Initializer.py


After successful completion,
two Dist and Sourse directories will appear.
Compiled GUI installer is in Dist. Initializer

To start it, you must give it the right to run.
sudo chmod -x Initializer