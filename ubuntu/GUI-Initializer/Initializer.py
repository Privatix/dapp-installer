import sys

from PyQt5 import QtWidgets

from sources.welcomePage import Configure


if __name__ == "__main__":
    app = QtWidgets.QApplication(sys.argv)
    MainWindow = QtWidgets.QMainWindow()
    ui = Configure(MainWindow)
    MainWindow.closeEvent = ui.UserEvent
    MainWindow.show()
    sys.exit(app.exec_())
