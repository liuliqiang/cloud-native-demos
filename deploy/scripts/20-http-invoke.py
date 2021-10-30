import json
import argparse

import requests


rst = {}

def record_target(target):
    if target in rst:
        rst[target] += 1
    else:
        rst[target] = 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Python http client test.')
    parser.add_argument('-time', dest='time', help='times to request http server')
    args = parser.parse_args()

    try:
        for i in xrange(int(args.time)):
            try:
                resp = requests.get("http://127.0.0.1:3333/ping/ping?say=hi")
            except Exception as e:
                print("failed to ping: {}".format(e))
            else:
                target = resp.text.split(":")[0]
                record_target(target)
                resp = None
    except KeyboardInterrupt:
        pass

    print(json.dumps(rst, indent=2))
        