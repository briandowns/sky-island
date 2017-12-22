# coding: utf-8

import json
import requests


class Encoder(json.JSONEncoder):
    def default(self, obj):
        if not isinstance(obj, Call):
            return super(Encoder, self).default(obj)
        return obj.__dict__


class Call(object):
    def __init__(self, url, call):
        self.url = url
        self.call = call


class Client(object):
    def __init__(self, host, port):
        self.endpoint = "http://{}:{}/api/v1/function".format(host, port)

    def function(self, url, call, full_resp=False):
        c = Call(url, call)
        res = requests.post(self.endpoint, json.dumps(c, cls=Encoder))
        if res.status_code == 200:
            data = json.loads(res.text)
            if full_resp:
                return data
            return data["data"]
