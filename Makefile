GO111MODULE=on

install:
	@go get

gen_example: install
	@protoc --go_out=plugins=grpc:./examples/ --gohttp_out=./examples/ -I examples ./examples/*.proto

test:
	@go test ./...

test_example:
	@go test ./examples/...

run_examples:
	@go run examples/main.go examples/greeter.pb.go examples/greeter.http.go

curl_google_api:
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > examples/google/api/annotations.proto 
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > examples/google/api/http.proto
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > testdata/google/api/annotations.proto 
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > testdata/google/api/http.proto

gen_test_grpc_go_files:
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/auth/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/hellostreamingworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/helloworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/httprule/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/routeguide/*.proto
