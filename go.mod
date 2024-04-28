// DON'T EDIT THIS!
//
// Codecrafters relies on this file being intact to run tests successfully. Any changes
// here will not reflect when CodeCrafters tests your code, and might even cause build
// failures.
//
// DON'T EDIT THIS!

module github.com/codecrafters-io/redis-starter-go

go 1.22

// for local testing
// replace github.com/codecrafters-io/redis-starter-go/token => ./token
// replace github.com/codecrafters-io/redis-starter-go/replication => ./replication

// for codecrafters test
replace github.com/codecrafters-io/redis-starter-go/token => ../token
replace github.com/codecrafters-io/redis-starter-go/replication => ../replication


require (
	github.com/codecrafters-io/redis-starter-go/replication v0.0.0-00010101000000-000000000000
	github.com/codecrafters-io/redis-starter-go/token v0.0.0-00010101000000-000000000000
)
