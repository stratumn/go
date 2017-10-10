docker rm -f $(docker ps -aq)
docker rmi $(docker images | grep dev-peer0.org1.example.com-pop-1.0 | awk "{print \$3}")