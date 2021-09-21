# grpc-echo


A demo for envoy xds and grpc

### Run

- grpc server
    
    ```
    # make docker
    # make up
    ```

- grpc client

    ```
    # go run cmd/client/client.go -port 3000 
    ```

- ads

    ```
    # go run cmd/xds/main.go -ads
    ```

- xds-grpc-client
  
    ```
    # make run-xds-grpc
    ```

### Envoy

#### Static proxy

```
# envoy -c deploy/envoy/00-static-envoy.yaml
```

test with grpc client

```
# make client
# .build/client -port 10000
2021/08/11 23:14:02  info from 77777: say:"hi"
# .build/client -port 10000
2021/08/11 23:14:05  info from 33333: say:"hi"
# .build/client -port 10000
2021/08/11 23:14:05  info from 55555: say:"hi"
# .build/client -port 10000   
2021/08/11 23:14:08  info from 55555: say:"hi"
```

#### xDS

```
# envoy -c deploy/envoy/12-lds-cds.yaml
```

test with grpc client

```
# make client
# .build/client -port 10000
2021/08/11 23:14:02  info from 77777: say:"hi"
# .build/client -port 10000
2021/08/11 23:14:05  info from 33333: say:"hi"
# .build/client -port 10000
2021/08/11 23:14:05  info from 55555: say:"hi"
# .build/client -port 10000   
2021/08/11 23:14:08  info from 55555: say:"hi"
```