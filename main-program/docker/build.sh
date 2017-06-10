echo "building..."
mkdir src
cd src
export GOPATH=/src
echo "getting sources..."
go get -u github.com/google/uuid
go get -u github.com/milak/mmq
go get -u github.com/milak/tools
ls
echo "compiling..."
export CGO_ENABLED=0
go build github.com/milak/mmq/mmq/mmq.go
echo "compile successful"