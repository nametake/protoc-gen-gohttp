GO111MODULE=on

install:
	@go get

gen_examples: install
	@protoc --go_out=./_examples/ --gohttp_out=./_examples/ --go_opt=paths=source_relative -I_examples ./_examples/*.proto

gen_pb:
	@protoc --go_out=./testdata/ --gohttp_out=./testdata/ --go_opt=paths=source_relative -I testdata ./testdata/**/*.proto

test: test_examples
	@go test ./...

test_examples:
	@cd _examples && go test

run_examples:
	@cd _examples && go run main.go greeter.pb.go greeter.http.go

curl_google_api:
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > _examples/google/api/annotations.proto
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto > testdata/google/api/annotations.proto
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > _examples/google/api/http.proto
	@curl https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto > testdata/google/api/http.proto
	@curl https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto> _examples/google/protobuf/any.proto
	@curl https://raw.githubusercontent.com/protocolbuffers/protobuf/master/src/google/protobuf/any.proto> testdata/google/protobuf/any.proto
