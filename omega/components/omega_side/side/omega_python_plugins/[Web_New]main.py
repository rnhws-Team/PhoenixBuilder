# 插件 开
from omega_side.python3_omega_sync import API
from omega_side.python3_omega_sync import frame as omega
from omega_side.python3_omega_sync.bootstrap import install_lib
from omega_side.python3_omega_sync.protocol import *
install_lib(lib_name="rich",lib_install_name="Rich")
install_lib(lib_name="schedule",lib_install_name="schedule")
install_lib(lib_name="PIL",lib_install_name="pillow")
install_lib(lib_name="flask",lib_install_name="flask")
install_lib(lib_name="psutil",lib_install_name="psutil")
import os,threading,json,sqlite3,base64,flask,logging,random,io,string,secrets,schedule,psutil
from PIL import Image,ImageDraw,ImageFont
import wnLogging as Logging

class MainCore (object):
    def __init__(self) -> None:
        self.AppConfig:dict = {}
        self.database = {}
        self.Log = Logging.Logging_Core(os.path.basename(__file__))
        self.ConfigFilePath:str = "AppData/AppConfig.json"
        self.TmpFilePath:str = "AppData/AppTmp.tmp"
        self.WebUserInitPath = "AppData/WebUserInit"
        if not os.path.exists(self.ConfigFilePath):
            self.Log.Error(f"{self.ConfigFilePath}，文件不存在,请确定下载到的文件完整!")
            input();exit(1)
        else:
            tmp_1 = json.load(open(file=self.ConfigFilePath,mode="r"))
            if len(tmp_1) == 0:
                self.Log.Error(f"{self.ConfigFilePath}，文件已损坏,请重新下载该文件!")
                input();exit(1)
            else:
                self.AppConfig = tmp_1
                tmp_2 = self.AppConfig["database_folder"]
            if not os.path.exists(tmp_2):
                os.makedirs(tmp_2)
            if not os.path.exists(f"{tmp_2}/AppDataBase_Main.db"):
                for i in self.AppConfig["databases"]:
                    tmp_1 = self.AppConfig["database_folder"]
                    self.Log.Info(f"创建/连接数据库文件,{tmp_1}/{i}")
                    if i == "AppDataBase_Web.db":
                        sqlite3.connect(f"{tmp_1}/{i}").cursor().execute('''CREATE TABLE main (
                                                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                                                        username TEXT,
                                                                        password TEXT
                                                                        );''')
                    else:
                        sqlite3.connect(f"{tmp_1}/{i}")
                    self.database[i] = sqlite3.connect(f"{tmp_1}/{i}")
        if not os.path.exists(self.WebUserInitPath):
            self.WebUserInit:bool = False
        else:
            self.WebUserInit:bool = True

    def get_omega_api(self,api:API):
        self.api=api
    def main (self):
        self.Tmp = Tmp(self.TmpFilePath)
        self.Web_Core = Web_Core()
        schedule.every().hour.do(self.Web_Core.usertokenlist_Clean)
        schedule.run_pending()
        omega.add_plugin(plugin=self.get_omega_api)
        omega.run(addr=None)
class Tmp (object):
    def __init__(self,TmpFilePath) -> None:
        if os.path.exists(f"{TmpFilePath}"):
            os.remove(f"{TmpFilePath}")
        self.TmpData = open(file=f"{TmpFilePath}",mode="wb+")
    def WriteJson (self,Json):
        self.TmpData.write(base64.b64encode(json.dumps(Json).encode()))
    def ReadJson (self):
        self.TmpData.seek(0)  # 指针回到文件头
        data = self.TmpData.read()
        data = base64.b64decode(data).decode()
        return json.loads(data)
        
class Web_Core (object):
    def __init__(self):
        self.WebConfig = {"host": "127.0.0.1", "port": 24012, "debug": True}
        self.WebApp = flask.Flask(
            "Web_NewWeb",
            static_folder=Core.AppConfig["static_folder"],
            template_folder=Core.AppConfig["template_folder"],
        )
        self.WebApp.secret_key = secrets.token_hex(16)
        log = logging.getLogger('werkzeug')
        self.session:dict = {}
        self.usertoken_12:dict = {}
        log.setLevel(logging.ERROR)
        self.Web()
        self.Api()
        
    def Web(self):
        @self.WebApp.route("/")
        @self.WebApp.route("/login")
        def login():
            if Core.WebUserInit == True:
                return flask.render_template("./NewWeb/Login")
            else:
                return flask.render_template("./NewWeb/LoginInit")
            
        @self.WebApp.route("/<token>/home")
        def home(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                        tmp_5 = v
                return flask.render_template("./NewWeb/Home",token=k,username=tmp_5)
            else:
                return flask.redirect('/login')

        @self.WebApp.route("/<token>/page/home")
        def PageHome(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                        tmp_5 = v
                return flask.render_template("./NewWeb/page/home",token=k,username=tmp_5,ip=self.WebConfig["host"],port=str(self.WebConfig["port"]),debug=str(self.WebConfig["debug"]))
            else:
                return flask.redirect('/login')

        @self.WebApp.route("/<token>/page/player")
        def PagePlayer(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                return flask.render_template("./NewWeb/page/player",token=k)
            else:
                return flask.redirect('/login')
    
    def Api(self):
        @self.WebApp.route('/api/login', methods=['POST'])
        def ApiLogin():
            self.username = str(flask.request.form.get('username'))
            self.password = str(flask.request.form.get('password'))
            if Core.WebUserInit == False:
                tmp_1 = Core.AppConfig["database_folder"]
                tmp_2 = sqlite3.connect(f"{tmp_1}/AppDataBase_Web.db")
                tmp_2.cursor().execute(f"INSERT INTO main (username, password) VALUES ('{self.username}', '{self.password}');")
                tmp_2.commit()
                open (file=f"{Core.WebUserInitPath}",mode="w").write(json.dumps({"username":self.username,"password":self.password}))
                Core.WebUserInit = True
                return flask.redirect('/login')
            if Core.WebUserInit == True:
                # 获取用户名、密码和验证码
                self.username = str(flask.request.form.get('username'))
                self.password = str(flask.request.form.get('password'))
                self.captcha = str(flask.request.form.get('captcha')).lower()
                tmp_1 = Core.AppConfig["database_folder"]
                tmp_2 = sqlite3.connect(f"{tmp_1}/AppDataBase_Web.db")
                # tmp_2.cursor().execute(f"SELECT main.usernma,main.password FROM main WHERE main.usernma = {self.username} AND main.password ={self.password};")
                # result = tmp_2.cursor().fetchone()
                tmp_3 = tmp_2.execute(f"SELECT * FROM main WHERE username = '{self.username}' AND password = '{self.password}';")
                result = tmp_3.fetchone()
                Core.Log.Info(f"Login:{result}")
                if result :
                    if self.captcha != self.session['captcha'].lower():
                        return flask.jsonify({'success': False, 'message': '验证码错误'})
                    else:
                        Core.Log.Info(f"{self.usertoken_12}")
                        if not self.username in self.usertoken_12.values():
                            tmp_4 = secrets.token_urlsafe()
                            self.usertoken_12[f"{tmp_4}"] = self.username
                        else:
                            for k, v in self.usertoken_12.items():
                                if v == self.username:
                                    tmp_4 = k
                        return flask.jsonify({'success': True, 'message': '登录成功','murl':f'/{tmp_4}/home'})
                else:
                    return flask.jsonify({'success': False, 'message': '账号或密码错误'})
        
        @self.WebApp.route('/api/captcha', methods=['GET'])
        def captcha():
            captcha_str, im = self.gen_captcha_image()
            b = io.BytesIO()
            im.save(b, 'PNG')
            b.seek(0)
            self.session['captcha'] = captcha_str
            Core.Log.Info(f"NewCaptcha:{captcha_str}")
            return flask.send_file(b, mimetype='image/png')

        @self.WebApp.route("/api/<token>/config/init",methods=["POST","GET"])
        def ApiConfigInit(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                return self.BuildInitJson(tmp_4)
            else:
                return flask.redirect('/login')

        @self.WebApp.route("/api/<token>/getsysteminfo",methods=["POST"])
        def GetSystemInfo(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                CPU = 0.0
                for i in psutil.cpu_percent(percpu=True):
                    CPU = CPU + i
                total = psutil.virtual_memory().total
                total = total/1024000000
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                return flask.jsonify({"cpupercent":f"{round(CPU/4,0)}%","cpucount":f"{psutil.cpu_count()}","virtualtotal":f"{int(float(total))}GB","virtualpercent":f"{psutil.virtual_memory().percent}%"})
            
            else:
                return flask.redirect('/login')
            
        @self.WebApp.route("/api/<token>/getgamemsg",methods=["POST"])
        def GetGameMsg(token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                game_msg = open(file="../logs/聊天记录.log",mode="r",encoding="utf-8").read()
                return flask.jsonify({"msg":f"{game_msg}"})
            
            else:
                return flask.redirect('/login')

        @self.WebApp.route("/api/<token>/getplayerjson")
        def GetPlayerJson (token):
            if not f"{token}" in self.usertoken_12:
                return flask.redirect('/login')
            elif f"{token}" in self.usertoken_12:
                for k, v in self.usertoken_12.items():
                    if v == self.username:
                        tmp_4 = k
                response=Core.api.do_get_players_list(cb=None)
                Form_Player = []
                js = 0
                for player in response:
                    js +=1
                    Form_Player.append({"id":js,"uuid":player.uuid,"username":player.name})
                return flask.jsonify({"code":0,"data":Form_Player})
            else:
                return flask.redirect('/login')

    def Run(self):
        self.WebApp.run(
            host=self.WebConfig["host"],
            port=self.WebConfig["port"],
            debug=self.WebConfig["debug"],
            use_reloader=False
        )

    def Start(self):
        threading.Thread(target=self.Run).start()

    def gen_rand_string(self,length):
        letters = string.ascii_lowercase
        return ''.join(random.choice(letters) for i in range(length))

    def gen_rand_color(self):
        return tuple(random.randint(0, 255) for i in range(3))

    def gen_captcha_image(self,width=120, height=40, length=4, fontsize=34):
        im = Image.new('RGB', (width, height), (255, 255, 255))
        draw = ImageDraw.Draw(im)
        captcha_str = self.gen_rand_string(length)
        font_file = 'arial.ttf'
        font = ImageFont.truetype(font_file, fontsize)
        draw.text((20, 0), captcha_str, font=font, fill=self.gen_rand_color())
        for i in range(10):
            x1, y1 = random.randint(0, width), random.randint(0, height)
            x2, y2 = random.randint(0, width), random.randint(0, height)
            draw.line([(x1, y1), (x2, y2)], fill=self.gen_rand_color(), width=2)
        for i in range(500):
            x, y = random.randint(0, width), random.randint(0, height)
            draw.point([x, y], fill=self.gen_rand_color())

        return captcha_str, im
    def usertokenlist_Clean (self):
        self.usertoken_12 = {}

    # def on_mc_packet(self,packet):
    #     if not packet["id"] in ["IDSetTime","IDActorEcent","IDLevelEvent","IDSetTitle","IDUpdateBlock","IDUpdateBlock",
    #                          "IDUpdateAttributes","IDLevelSoundEvent","IDMobEffect","IDSetActorMotion","IDMoveActorDelta",
    #                          "IDCommandOutput","IDGameRulesChanged","IDSetActorData","IDActorEvent"]:
    #         if packet["id"] == "IDPlayerList":
    #             print(f"类型{packet.id}/{packet}")


    def BuildInitJson(self,token):
        return {
                "homeInfo": {
                    "title": "首页",
                    "href": f"/{token}/page/home"
                },
                "logoInfo": {
                    "title": "New_Web",
                    "href": ""
                },
                "menuInfo": [
                    {
                    "title": "NewWeb",
                    "icon": "fa fa-window-maximize",
                    "href": "",
                    "target": "_self",
                    "child": [
                        {
                        "title": "在线列表",
                        "href": f"/{token}/page/player",
                        "icon": "fa fa-window-maximize",
                        "target": "_self"
                        },
                        {
                        "title": "系统设置",
                        "href": "./layuimini-v2/page/setting.html",
                        "icon": "fa fa-gears",
                        "target": "_self"
                        }
                    ]}
                ]}

Core = MainCore()
Core.main()
Core.Web_Core.Start()   

