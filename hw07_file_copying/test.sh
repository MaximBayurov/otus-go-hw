#!/usr/bin/env bash
set -xeuo pipefail

TO=out.txt
FROM=testdata/input.txt

go build -o go-cp

./go-cp -from $FROM -to $TO
cmp $TO testdata/out_offset0_limit0.txt

./go-cp -from $FROM -to $TO -limit 10
cmp $TO testdata/out_offset0_limit10.txt

./go-cp -from $FROM -to $TO -limit 1000
cmp $TO testdata/out_offset0_limit1000.txt

./go-cp -from $FROM -to $TO -limit 10000
cmp $TO testdata/out_offset0_limit10000.txt

./go-cp -from $FROM -to $TO -offset 100 -limit 1000
cmp $TO testdata/out_offset100_limit1000.txt

./go-cp -from $FROM -to $TO -offset 6000 -limit 1000
cmp $TO testdata/out_offset6000_limit1000.txt

rm -f $TO
./go-cp -from $FROM -to $TO -offset 100000
if [ -f $TO ]; then
  exit 0
fi

rm -f go-cp out.txt
echo "PASS"
