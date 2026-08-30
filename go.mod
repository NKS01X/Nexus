module github.com/razorpay/aegis

replace github.com/razorpay/nexus-mcp => ./mcp-lib

go 1.22

require (
	github.com/jackc/pgx/v5 v5.5.0
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/razorpay/nexus-mcp v0.0.0

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.13.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)
