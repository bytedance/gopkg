CPU=1
T=.
B=.
C=1

test:
	go test -coverprofile=cover.out -v ./... -count $(C) -run $(T)

bench:
	go test -v -cpuprofile cpu.pprof -cpu=$(CPU) -run=NOTEST -bench=$(B) -benchmem -benchtime=100000x -count=10 -timeout=60m > y.txt && benchstat y.txt

bench-itoa:
	go test -cpu=1 -run=NOTEST -bench=$(B) -benchtime=100000x -benchmem -count=50 -timeout=60m -run BenchmarkContains100Hits^ > 1.txt
	benchstat 1.txt

race:
	go test -cpu=12 -race -run=NOTEST -bench=. -benchtime=100000x -benchmem -count=10 -timeout=60m
