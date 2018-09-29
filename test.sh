#!/bin/sh

port=54323
pod=`kubectl get pod -l draft=cryptogen -o name | awk -F'/' '{ print $2 }'`

echo "Pod $pod"

kubectl port-forward $pod $port:5000 2>&1 > /dev/null &
pid=$!
sleep 1

echo "pid: $pid"

echo "health check:"
curl -sS http://localhost:$port/health
echo


echo "api test:"
res=`curl -sS -XPOST \
    -d'{"PeerOrgs": [{"Name": "Org1", "Domain": "org1.example.com", "Template": {"Count": 1}, "Users": {"Count": 1}}]}' \
    -H"Content-type: application/json" \
    http://localhost:$port/crypto-assets`

echo $res

jobID=`echo $res | jq -r '.JobID'`
cmd="curl -sS http://localhost:$port/status/$jobID"

echo "$cmd"
$cmd && echo

sleep 3
echo "$cmd"
$cmd && echo


kill $pid
