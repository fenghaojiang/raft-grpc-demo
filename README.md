# Raft GRPC Demo

raft realization demo using grpc

## Start Raft RegisterCenter

registerCenter --- 127.0.0.1:50000

```shell
cd ./register
go build -o registerCenter
./registerCenter
```

## Start your own cluster

```shell
go build -o raft-demo
```

```shell
./registerCenter  

./raft-demo --svc 127.0.0.1:51000 --id node1 --data data/node1 --raft 127.0.0.1:52000 --service_join 127.0.0.1:50000

./raft-demo --svc 127.0.0.1:51001 --id node2 --data data/node2 --raft 127.0.0.1:52001 --join 127.0.0.1:51000 --service_join 127.0.0.1:50000

./raft-demo --svc 127.0.0.1:51002 --id node3 --data data/node3 --raft 127.0.0.1:52002 --join 127.0.0.1:51000 --service_join 127.0.0.1:50000
```

## Reference

https://github.com/Jille/raft-grpc-example
<br>
https://github.com/hanj4096/raftdb
<br>
https://github.com/HelKim/raft-demo
<br>

## Go Report Card Score

```
Grade: A (89.2%)
Files: 12
Issues: 5
gofmt: 100%
go_vet: 100%
gocyclo: 91%
golint: 66%
license: 0%
ineffassign: 100%
misspell: 0%
```
