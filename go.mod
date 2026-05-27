module ipcheck

go 1.26.3

require (
	github.com/miekg/dns v1.1.72
	github.com/quic-go/quic-go v0.59.1
	golang.org/x/net v0.55.0
)

require (
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/mobile v0.0.0-20260520154334-0e4426e1883d // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
)

tool (
	golang.org/x/mobile/cmd/gobind
	golang.org/x/mobile/cmd/gomobile
)
