module example

go 1.24.0

toolchain go1.24.9

replace github.com/ryomak/gopdf => ../..

require github.com/ryomak/gopdf v0.0.0

require (
	golang.org/x/image v0.32.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)
