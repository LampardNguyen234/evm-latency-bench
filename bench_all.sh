
#!/bin/bash

txcount=100
plot_dir="plots"

echo "Benchmarking Arbitrum..."
echo "Getting the baseline"
cp arb.env .env && go run . bench resp-time --count 20
echo "Total time"
go run . bench --txcount $txcount --plot-prefix "arb" --plot-dir=$plot_dir --plot
echo "DONE Arbitrum"
echo ""
echo ""

echo "Benchmarking Base..."
echo "Getting the baseline"
cp base.env .env && go run . bench resp-time --count 20

echo "Total time"
go run . bench --txcount $txcount --plot-prefix "base" --plot-dir=$plot_dir --plot
echo "DONE Base"
echo ""
echo ""

echo "Benchmarking MegaETH..."
echo "Getting the baseline"
cp mega.env .env && go run . bench resp-time --count 20

echo "Total time async"
go run . bench --txcount $txcount --plot-prefix "mega_async" --plot-dir=$plot_dir --plot

echo "Total time compare"
go run . bench compare --txcount $txcount --plot-prefix "mega" --plot-dir=$plot_dir --plot
echo "DONE MegaETH"
echo ""
echo ""

echo "Benchmarking Optimism..."
echo "Getting the baseline"
cp op.env .env && go run . bench resp-time --count 20
echo "Total time"
go run . bench --txcount $txcount --plot-prefix "op" --plot-dir=$plot_dir --plot

echo "DONE Optimism"
echo ""
echo ""

echo "Benchmarking RISE..."
echo "Getting the baseline"
cp rise.env .env && go run . bench resp-time --count 20
echo "Total time async"
go run . bench --txcount $txcount --plot-prefix "rise_async" --plot-dir=$plot_dir --plot
echo "Total time compare"
go run . bench compare --txcount $txcount --plot-prefix "rise" --plot-dir=$plot_dir --plot

echo "DONE Rise"
echo ""
echo ""
