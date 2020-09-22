#!/bin/sh

for f in bin/faasd*; do shasum -a 256 $f > $f.sha256; done
