module github.com/stefan-muehlebach/adatft

go 1.24.4

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
)
