import rich,datetime
class Logging_Core (object):
    def __init__(self,File_Name) -> None:
        self.File_Name = str(File_Name)
    def Debug (self,debug):
        debug = f"\033[33m[{self.File_Name}][{datetime.datetime.now().strftime('%Y-%m-%d-%H:%M:%S')}][Debug]{debug}\033[0m"
        print(debug)

    def Info (self,info):
        info = f"\033[38m[{self.File_Name}][{datetime.datetime.now().strftime('%Y-%m-%d-%H:%M:%S')}][Info]{info}\033[0m"
        print(info)
    def Error (self):
        error = f"\033[31m[{self.File_Name}][{datetime.datetime.now().strftime('%Y-%m-%d-%H:%M:%S')}][Error]{error}\033[0m"
        print(error)

