import time
import requests

for i in xrange(100000000):
    try:
        resp = requests.get("http://127.0.0.1:10000/ping/ping?say=hi")
    except Exception as e:
        print("failed to ping: {}".format(e))
    print(resp.text)
    time.sleep(0.5)    
