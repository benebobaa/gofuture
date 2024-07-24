
test:
	go test -v 

trace:
	go test -v -race

bench:
	go test -bench=. -benchmem

bench5:
	go test -bench=. -benchmem -benchtime=5s -count=5

profile:
	go test -bench=. -benchtime=5s -cpuprofile=cpu.out

