# ADATFT

`adatft` is a Go library for the RaspberryPi and gives you access to the TFT display from Adafruit.
It does even more! It allows you to draw on the display with a proven Go graphics package: `gg`.
And it does this so fast, on RaspberryPi 2 you can even display animations with approx. 25 fps.

## Installation

Since `adatft` consists of a Go package on one side but has also a part written in C responsible for the hardware access,
you cannot just install the package with `git get` - a C compiler and a number of libraries are needed.

