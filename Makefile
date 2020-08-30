GO111MODULE=on

install:
	@go get

gen_examples: install
	@protoc --go_out=plugins=grpc:./_examples/ --gohttp_out=./_examples/ --go_opt=paths=source_relative -I_examples ./_examples/*.proto

test:
	@go test ./... ./_examples/

test_examples:
	@go test ./_examples/

run_examples:
	@go run _examples/main.go _examples/greeter.pb.go _examples/greeter.http.go

update_libs:
	@go get -u  github.com/golang/protobuf google.golang.org/grpc
	@go mod tidy

curl_google_api:
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > _examples/google/api/annotations.proto 
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > _examples/google/api/http.proto
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > testdata/google/api/annotations.proto 
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > testdata/google/api/http.proto

gen_test_grpc_go_files:
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/auth/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/hellostreamingworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/helloworld/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/httprule/*.proto
	@protoc --go_out=plugins=grpc:./testdata/ -I testdata ./testdata/routeguide/*.proto
