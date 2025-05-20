#!/bin/bash

txcount=$1
# shellcheck disable=SC1073
if [ -z $txcount ]; then
  txcount=100;
fi
echo "txcount: ${txcount}"
plot_dir="plots"

echo "Benchmarking MegaETH..."
cp mega.env .env
#echo "Total time async"
#go run . bench --txcount $txcount --poll-interval "1ms" --plot-prefix "mega_async" --plot-dir=$plot_dir --plot
#
#echo "Total time sync"
#go run . bench --txcount $txcount --mode "sync" --plot-prefix "mega_sync" --plot-dir=$plot_dir --plot

echo "Total time compare"
go run . bench compare --txcount $txcount --poll-interval "10ms" --plot-prefix "mega" --plot-dir=$plot_dir --plot
echo "DONE MegaETH"
echo ""
echo ""

echo "Benchmarking RISE..."
cp rise.env .env
#echo "Total time async"
#go run . bench --txcount $txcount --poll-interval "1ms" --plot-prefix "rise_async" --plot-dir=$plot_dir --plot
#echo "Total time sync"
#go run . bench --txcount $txcount --mode "sync" --plot-prefix "rise_sync" --plot-dir=$plot_dir --plot

echo "Total time compare"
go run . bench compare --txcount $txcount --poll-interval "10ms" --plot-prefix "rise" --plot-dir=$plot_dir --plot

echo "DONE Rise"
echo ""
echo ""
