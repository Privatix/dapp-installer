
"""
        GUI Initializer

        Version 0.2.10
"""
import sys
from os import path
from PyQt5 import QtWidgets

from sources.welcomePage import Configure

if __name__ == "__main__":
    reldirname = path.dirname(path.abspath(__file__)) + '/'
    app = QtWidgets.QApplication(sys.argv)
    main_obj = QtWidgets.QMainWindow()
    ui = Configure(main_obj, reldirname)
    main_obj.closeEvent = ui.UserEvent
    main_obj.show()
    sys.exit(app.exec_())
