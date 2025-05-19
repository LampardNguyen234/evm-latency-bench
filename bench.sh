#!/bin/bash

txcount=10
plot_dir="plots"
name=$1

echo "Benchmarking $name..."
echo "Getting the baseline"
cp $name.env .env && go run . bench resp-time --count $txcount

echo "Total time"
go run . bench --txcount $txcount --plot-prefix "$name" --plot-dir=$plot_dir --plot
echo "DONE $name"
echo ""
echo ""
