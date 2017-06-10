echo "building..."
ls
mkdir wokspace
ls
cd wokspace
ls
export GOPATH=/wokspace
echo "getting sources..."
go get -u github.com/google/uuid
go get -u github.com/milak/mmq
go get -u github.com/milak/tools
ls
echo "compiling..."
export CGO_ENABLED=0
go build src/github.com/milak/mmq/mmq/mmq.go
echo "compile successful"