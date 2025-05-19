#!/bin/bash

txcount=10
plot_dir="plots"
name=$1

cp $name.env .env
echo "Benchmarking Arbitrum..."
echo "interval=10ms"
go run . bench receiptcount --txcount $txcount --plot-prefix "$name-count-receipt-10ms" --plot-dir=$plot_dir --plot
echo "interval=50ms"
go run . bench receiptcount --txcount $txcount --plot-prefix "$name-count-receipt-50ms" --plot-dir=$plot_dir --plot
echo "interval=100ms"
go run . bench receiptcount --txcount $txcount --plot-prefix "$name-count-receipt-100ms" --plot-dir=$plot_dir --plot
