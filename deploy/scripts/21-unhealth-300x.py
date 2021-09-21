import requests


for i in xrange(7):
    requests.get("http://localhost:{}/change_health".format(3000 + i+3))