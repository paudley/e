test:
	@go test -failfast -race -coverprofile=tests_cover_profile -cpuprofile=tests_cpu_profile -mutexprofile=tests_mutex_profile -memprofile=tests_mem_profile -blockprofile=tests_block_profile -shuffle=on -vet=all -parallel 8 -covermode=atomic ./... -v
	@rm e.test
	@go tool pprof -text tests_cpu_profile > tests_cpu_profile.txt
	@go tool pprof -text tests_mem_profile > tests_mem_profile.txt
	@go tool pprof -text tests_mutex_profile > tests_mutex_profile.txt
	@go tool pprof -text tests_block_profile > tests_block_profile.txt
	@go tool cover -func tests_cover_profile -o tests_cover_profile.txt
	@go tool cover -html tests_cover_profile -o tests_cover_profile.html
	@golangci-lint run

clean:
	rm tests_*_profile* e.test
