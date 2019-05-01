import json
import urllib
import socket

class Context(object):
    def Ctx(self):
        msg.echo("context")
    def Cmd(self, msg):
        msg.echo("command %s", dir(self))

    def config(self, msg):
        for k, v in self.configs.iteritems():
            msg.append("key", k)
            msg.append("value", str(json.dumps(v)))
        msg.table()
        msg.log("fuck %s", msg.meta["result"])

    def Cap(self):
        msg.echo("cache")

class Message(object):

    code = 0
    @classmethod
    def _incr_count(cls):
        cls.code += 1
        return cls.code

    def __init__(self, m=None, target=None, remote_code=-1, message=None, meta={}):
        self.code = self._incr_count()
        self.remote_code = remote_code
        self.message = m.code if m else message
        self.target = target
        self.meta = meta

    def detail(self, *argv):
        detail = self.meta.get("detail", [])
        if argv:
            try:
                i = int(argv[0])
                if 0 <= i and i < len(detail):
                    return detail[i]
                else:
                    return ""
            except:
                self.meta["detail"] = argv
        return self.meta.get("detail") or []

    def detaili(self, *argv):
        v = self.detail(*argv)
        try:
            return int(v)
        except:
            return 0

    def option(self, key=None, value=None):
        if key:
            msg = self
            while msg:
                if key in msg.meta["option"]:
                    return msg.meta[key]
                msg = msg.message
            return []
        return self.meta.get("option", [])

    def append(self, key=None, value=None):
        if value is not None:
            self.meta[key] = self.meta.get(key, [])
            self.meta[key].append(value)
            if key not in self.meta.get("append", []):
                self.meta["append"] = self.meta.get("append", [])
                self.meta["append"].append(key)

        if key is not None:
            return self.meta.get(key, [""])[0]

        return self.meta.get("append", [])

    def result(self, s=None):
        if s:
            self.meta["result"] = self.meta.get("result", [])
            self.meta["result"].append(s)
        return self.meta.get("result", [])

    def table(self):
        nrow = len(self.meta[self.meta["append"][0]])
        ncol = len(self.meta["append"])

        for j, k in enumerate(self.meta["append"]):
            self.echo(k)
            self.echo(" " if j < ncol-1 else "\n")
        for i in range(nrow):
            for j, k in enumerate(self.meta["append"]):
                self.echo(self.meta[k][i])
                self.echo(" " if j < ncol-1 else "\n")

    def echo(self, s, *argv):
        if len(argv) == 0:
            self.result(s if len(argv) == 0 else s % argv)

    def Conf(self, key=None, value=None):
        if value is not None:
            self.target.configs[key] = value
        if key is not None:
            return self.target.configs.get(key, "")
        return self.target.configs

    def log(self, s, *argv):
        print s % argv

class SSH(Context):
    def __init__(self, m, target=None, name="demo", host="127.0.0.1", port=9090):
        self.target = target or self
        self.m = m
        self.host = host
        self.port = port
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.socket.connect((host, port))
        msg = Message(m)
        msg.detail("recv", "add", name, "worker")
        self.sendDetail(msg, self.recvResult, self)
        self.recvLoop()

    sendMM = {}
    def sendDetail(self, msg, fun=None, obj=None):
        msg.recvFun = fun
        msg.recvObj = obj
        self.sendMM[msg.code] = msg
        self.socket.send("code: %s\n" % msg.code)
        data = ["detail: %s" % urllib.quote(d) for d in msg.detail()]
        for k, option in enumerate(msg.option()):
            for i, v in enumerate(option):
                data.append("%s: %s" % (urllib.quote(k) ,urllib.quote(v)))
        data.append("\n")
        self.socket.send("\n".join(data))
        msg.log("send: %s %s", msg.code, data)

    def recvResult(self, msg):
        msg.log("echo: %s", msg.result())

    def recvLoop(self):
        m = self.m
        self.file = self.socket.makefile()

        meta = {}
        while True:
            l = self.file.readline()
            if l == "\n":
                m.log("recv: %s %s", meta.get("remote_code"), meta)
                if "detail" in meta:
                    msg = Message(m, self.target, remote_code=meta.get("remote_code"), meta=meta)
                    self.recvDetail(msg)
                else:
                    msg = self.sendMM[int(meta.get("remote_code"))]
                    msg.meta.update(meta)
                    msg.recvFun and msg.recvFun(msg)
                    msg.recvObj and msg.recvObj.recvResult(msg)
                meta = {}
                continue

            v = l.rstrip("\n").split(": ")
            if v[0] == "code":
                meta["remote_code"] = v[1]
            elif v[0] == "detail":
                meta["detail"] = meta.get("detail", [])
                meta["detail"].append(v[1])
            elif v[0] == "result":
                meta["result"] = meta.get("result", [])
                meta["result"].append(v[1])
            else:
                if v[0] == "remote_code":
                    continue
                if v[0] not in meta:
                    if "detail" in meta:
                        meta["option"] = meta.get("option", [])
                        meta["option"].append(v[0])
                    else:
                        meta["append"] = meta.get("append", [])
                        meta["append"].append(v[0])
                meta[v[0]] = meta.get(v[0], [])
                meta[v[0]].append(v[1])

    def pwd(self, msg):
        msg.echo("shaoying")

    def recvDetail(self, msg):
        if hasattr(self.target, msg.detail(1)):
            getattr(self.target, msg.detail(1))(msg)
        self.sendResult(msg)

    def sendResult(self, msg):
        self.socket.send("code: %s\n" % msg.remote_code)
        data = ["result: %s" % urllib.quote(d) for d in msg.result()]
        for i, k in enumerate(msg.append()):
            for i, v in enumerate(msg.meta[k]):
                data.append("%s: %s" % (urllib.quote(k) ,urllib.quote(v)))
        data.append("\n")
        self.socket.send("\n".join(data))
        msg.log("send: %s %s", msg.remote_code, data)

try:
    import RPi.GPIO as GPIO
except:
    GPIO = Context()
    pass

class RPI(Context):
    def __init__(self):
        self.configs = {
                "setmode": False,
                }

    def gpio(self, msg, *argv):
        if msg.detail(2) == "init":
            GPIO.setmode(GPIO.BCM)
        elif msg.detail(2) == "out":
            if not msg.Conf("setmode"):
                msg.log("default setmode")
                msg.Conf("setmode", True)
                GPIO.setmode(GPIO.BCM)
            if msg.Conf(msg.detail(3)) != GPIO.OUT:
                msg.log("default output")
                msg.Conf(msg.detail(3), GPIO.OUT)
                GPIO.setup(msg.detaili(3), GPIO.OUT)
            GPIO.output(msg.detaili(3), msg.detaili(4))
        elif msg.detail(2) == "exit":
            GPIO.cleanup()

import sys
SSH(Message(), RPI(),
        sys.argv[1] if len(sys.argv) > 1 else "demo",
        sys.argv[2] if len(sys.argv) > 2 else "127.0.0.1")
