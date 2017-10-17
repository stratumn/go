#!/bin/bash
#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
# Exit on first error
set -e

# don't rewrite paths for Windows Git Bash users
export MSYS_NO_PATHCONV=1
export BASE_DIR=$GOPATH/src/github.com/stratumn/sdk/fabricstore/integration

starttime=$(date +%s)

if [ ! -d ~/.hfc-key-store/ ]; then
	mkdir ~/.hfc-key-store/
fi
cp $BASE_DIR/creds/* ~/.hfc-key-store/
# launch network; create channel and join peer to channel
./start.sh

# Now launch the CLI container in order to install and instantiate chaincode
docker-compose -f $BASE_DIR/docker-compose.yml up --build -d cli

docker exec cli go get github.com/stratumn/sdk/cs

docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" cli peer chaincode install -n pop -v 1.0 -p github.com/pop

docker exec -e "CORE_PEER_LOCALMSPID=Org1MSP" -e "CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp" cli peer chaincode instantiate -o orderer.example.com:7050 -C mychannel -n pop -v 1.0 -c '{"Args":[]}' -P "OR ('Org1MSP.member','Org2MSP.member')"

printf "\nTotal execution time : $(($(date +%s) - starttime)) secs ...\n\n"
