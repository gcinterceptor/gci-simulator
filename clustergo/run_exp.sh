#!/bin/sh

set -x

DURATIONMS=7200000
BIG_OMEGA=0.003
MU=0.0036
LITTLE_OMEGA=0.0001
NRUNS=50  # number of repetitions

# 1 replica
cd inputgen/;
go run main.go --duration=${DURATIONMS} --r=${NRUNS} --mu=${MU} --bigOmega=${BIG_OMEGA} --littleOmega=${LITTLE_OMEGA}
cd ..
for i in `seq 1 ${NRUNS}`
do
    go run main.go --rate=1 --warmup=0 --d=${DURATIONMS}ms --i=inputgen/input_$i.csv > r1_$i.out;
done

cat *.out | grep PCP | cut -d' ' -f1 | cut -d: -f2 > sim_1.pcp
cat *.out | grep PVN | cut -d' ' -f1 | cut -d: -f2 > sim_1.pvn

zip sim_1_out.zip r1_*.out
rm *.out

echo -e "\n ### REPLICA: PCP ###\n"
awk -v ORS=, '{ print $0 }' sim_1.pcp
echo -e "\n ###### \n"

echo -e "\n ### REPLICA: PVN ###\n"
awk -v ORS=, '{ print $0 }' sim_1.pvn
echo -e "\n ###### \n"