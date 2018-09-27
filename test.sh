#!/bin/sh

addr='localhost:5000'

rm -rf tmp/_cryptogen

make clean build
bin/cryptogen --addr $addr &
pid=$!

sleep 1

data='{"PeerOrgs": [{"Name": "Org1", "Domain": "org1.example.com", "Template": {"Count": 1}, "Users": {"Count": 1}}]}'
curl -v -XPOST \
    -d"$data" \
    -H"Content-type: application/json" \
    http://$addr/crypto-assets
echo

sleep 3

kill $pid
