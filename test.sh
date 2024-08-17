cd src
go get
go build -buildvcs=false
cd ../docker
cp ../src/noise-test .
docker build -t test . 
