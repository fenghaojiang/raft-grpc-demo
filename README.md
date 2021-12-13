# Raft GRPC Demo

raft realization demo using grpc   

## Start Raft RegisterCenter

registerCenter --- localhost:50000

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
./raft-demo --svc localhost:51000 --id node1 --data data/node1 --raft localhost:52000 --service_join localhost:50000

./raft-demo --svc localhost:51001 --id node2 --data data/node2 --raft localhost:52001 --join localhost:51000 --service_join localhost:50000

./raft-demo --svc localhost:51002 --id node3 --data data/node3 --raft localhost:52002 --join localhost:51000 --service_join localhost:50000
```

## Reference

https://github.com/Jille/raft-grpc-example
<br>
https://github.com/hanj4096/raftdb
<br>
https://github.com/HelKim/raft-demo
<br>
