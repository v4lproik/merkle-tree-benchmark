# merkle-tree
This is an attempt at demonstrating how a basic merkle tree library works and how it can be optimised using Go.
This merkle tree extends the most basic operations including building a tree, verifying that a leaf exist within the tree and so on.

The project has been coded with benchmarks in mind meaning that memory optimisation happened along the way. You can see the ```max-goroutine``` option is allowed to be specified as well as ```reuse-buffer-allocation``` performance options.
```errGroup``` and ```sync.Pool``` patterns have been introduced when building and verifying a tree in order to speed up the process of such actions as well as paying attention to any operation that could significantly increase the allocs/op such as "append".

## Config
```
merkle-tree:
  log:
    verbosity-level: "debug"
  hash: "sha256"
  performance:
    max-goroutine: 100000
    reuse-buffer-allocation: true
  data:
    - value1
    - value2
    - value3
    - value4
    - value5
    - value6
    - ...
```
There's a few options that allow you to configure the tree. It's pretty self-explanatory so I won't go into details here.
## Build
```
make build
GO111MODULE=on go build \
        -ldflags="-X 'main.GitSHA1=d684e56909f3033ae14140bb61840c6aaee8346b' -o merkle-tree main.go
```
## Run from CLI
```
# build the tree
./merkle-tree -c etc/conf.yml build
```
## Tests with race condition (+ coverage)
```
❯ make test
GO111MODULE=on go fmt ./...
GO111MODULE=on go test -race -covermode=atomic ./...
?   	github.com/v4lproik/merkle-tree	[no test files]
?   	github.com/v4lproik/merkle-tree/cmd	[no test files]
ok  	github.com/v4lproik/merkle-tree/pkg	(cached)	coverage: 81.8% of statements
```
## Benchmark
```
❯ make bench
go test -bench=. ./... -benchtime=5s -benchmem -count=6 -run=^#
goos: darwin
goarch: arm64
pkg: github.com/v4lproik/merkle-tree/pkg
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  536575	     10949 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  535962	     10974 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  540540	     10864 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  537912	     10919 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  540764	     10880 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000-10       	  542878	     10845 ns/op	    1299 B/op	      44 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38337	    149103 ns/op	    3834 B/op	     116 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38142	    150421 ns/op	    3843 B/op	     116 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38715	    148835 ns/op	    3817 B/op	     115 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38420	    150789 ns/op	    3831 B/op	     116 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38358	    150348 ns/op	    3833 B/op	     116 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N100000-10     	   38125	    149492 ns/op	    3844 B/op	     116 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1417	   3651518 ns/op	  476720 B/op	   12081 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1508	   3578604 ns/op	  448101 B/op	   11357 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1396	   3700686 ns/op	  483853 B/op	   12261 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1585	   3818895 ns/op	  426452 B/op	   10809 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1357	   3875475 ns/op	  497693 B/op	   12611 allocs/op
BenchmarkMerkleTreeBuilder_Verify_N1000000-10    	    1243	   4202731 ns/op	  543116 B/op	   13760 allocs/op
PASS
ok  	github.com/v4lproik/merkle-tree/pkg	392.531s
```
## More commands and tooling in the Makefile

## Improvements
- More algorithms being supported, as of right now, only sha256 is supported and change few algorithms that are bound to fixed array length  cf. TODO AI(Joel)
- Run everything in a CI (especially with benchmark as we want merkle tree to be highly performant)
- Dockerisation to be able to have env parity and test the library on different systems
- Have a unified interface for hash.Hash and HashPool as the code is slightly redundant when buffer reutilisation is activated
- Adding more actions to the merkle tree
~~- Sort mechanism (OpenZeppelin compatibility for instance)~~
