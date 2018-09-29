#!/bin/sh

pod=`kubectl get pod -l draft=cryptogen -o name | awk -F'/' '{ print $2 }'`

echo "Pod $pod"

if [ -f .test.pid ]; then
    pid=`cat .test.pid`
    echo "killing $pid"
    kill $pid
    rm .test.pid
fi

kubectl port-forward $pod 5000 &
pid=$!
if [ $? -eq 0 ]; then
    echo $pid > .test.pid
    sleep 1
fi

echo "pid: $pid"

echo "health check:"
curl -sS http://localhost:5000/health
echo

echo "api test:"
curl -sS -XPOST \
    -d'{"PeerOrgs": [{"Name": "Org1", "Domain": "org1.example.com", "Template": {"Count": 1}, "Users": {"Count": 1}}]}' \
    -H"Content-type: application/json" \
    http://localhost:5000/crypto-assets | jq -r '.TaskID' > .taskID

echo

taskid=`cat .taskID`
cmd="curl -sS http://localhost:5000/task/$taskid"
echo "$cmd"
$cmd | jq
