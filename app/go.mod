module github.com/codecrafters-io/redis-starter-go/app

go 1.22

replace github.com/codecrafters-io/redis-starter-go/token => ../token

replace github.com/codecrafters-io/redis-starter-go/replication => ../replication

require (
	github.com/codecrafters-io/redis-starter-go/replication v0.0.0-00010101000000-000000000000
	github.com/codecrafters-io/redis-starter-go/token v0.0.0-00010101000000-000000000000
)
