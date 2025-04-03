module github.com/stefan-muehlebach/adatft

go 1.24.2

replace github.com/stefan-muehlebach/gg => ../gg

require (
	github.com/stefan-muehlebach/gg v0.0.0-00010101000000-000000000000
	golang.org/x/image v0.25.0
	periph.io/x/conn/v3 v3.7.2
	periph.io/x/host/v3 v3.8.4
)

require (
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	periph.io/x/cmd v0.0.0-20250320001559-905e30669e74 // indirect
	periph.io/x/d2xx v0.1.1 // indirect
	periph.io/x/devices/v3 v3.7.4 // indirect
)
