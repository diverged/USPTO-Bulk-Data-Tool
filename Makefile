BINARY_NAME=usptgo

build:
	go build -o $(BINARY_NAME) ./cmd

.PHONY: build


run:
	go build -o $(BINARY_NAME) ./cmd
	./$(BINARY_NAME)
	
.PHONY: run

pprof:
	mkdir -p data/profiling/benchmarks/$(dir)
	go tool pprof -top $(BINARY_NAME) data/profiling/cpu.pprof > data/profiling/benchmarks/$(dir)/cpupprof.txt
	go tool pprof -png $(BINARY_NAME) data/profiling/cpu.pprof > data/profiling/benchmarks/$(dir)/cpupprof.png
	go tool pprof -top $(BINARY_NAME) data/profiling/mem.pprof > data/profiling/benchmarks/$(dir)/mempprof.txt
	go tool pprof -png $(BINARY_NAME) data/profiling/mem.pprof > data/profiling/benchmarks/$(dir)/mempprof.png
.PHONY: pprof


test:
	go test -v ./...

.PHONY: test