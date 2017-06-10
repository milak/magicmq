echo "installing go..."
apt-get install golang
echo "building..."
export CGO_ENABLED=0
mkdir src
export GOPATH=/src
echo "getting sources..."
go get -u github.com/google/uuid
go get -u github.com/milak/mmq
go get -u github.com/milak/tools
echo "compiling..."
#go build src/github.com/milak/mmq/mmq/mmq.go
#sudo docker build --rm -f docker/Dockerfile -t magicmq .
#rm docker/magicmq
echo "compile successful"