#!/bin/bash

set -x
set -e

LITTLE_OMEGA=${LITTLE_OMEGA:=0.0001}
BIG_OMEGA=${BIG_OMEGA:=0.003}
MU=${MU:=0.0036}
DURATIONMS=${DURATIONMS:=1200000}
NRUNS=${NRUNS:=5}  # number of repetitions
NREPLICAS=${NREPLICAS:=1} # comma-separated number of replicas (int32)
HT=${HT:=-1} # Hedging threshold (float64, default -1 and means no hedging)
HEDGE_CANCELLATION=${HEDGE_CANCELLATION:=true} # Whether to cancel hedge requests
ENABLE_CCT=${ENABLE_CCT:=true} # Whether to enable CCT (true or false)

for n in ${NREPLICAS//,/ }
do
    for run in `seq 1 ${NRUNS}`
    do
        echo -e "\n\n## Starting run $n : ${run} ##\n\n"
        cd inputgen/;
        go run main.go --duration=${DURATIONMS} --r=$n --mu=${MU} --bigOmega=${BIG_OMEGA} --littleOmega=${LITTLE_OMEGA} --cct=${ENABLE_CCT}
        cd ..
        input=""
        for i in `seq 1 ${n}`
        do
            input+="inputgen/input_${i}.csv,"
        done
        go run main.go --rate=1 --warmup=0 --d=$(( DURATIONMS ))ms --ht=${HT} --hedge-cancellation=${HEDGE_CANCELLATION} --cct=${ENABLE_CCT} --i=$input > r${n}_${run}.out        echo -e "\n\n## Finished run $n : ${run} ##\n\n"
    done
    cat *.out | grep DURATION | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.dur
    cat *.out | grep PCP | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pcp
    cat *.out | grep PVN | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pvn
    cat *.out | grep NUM_PROC_SUCC | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.succ
    cat *.out | grep NUM_PROC_FAILED | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.fail
    cat *.out | grep HEDGED | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.hedged
    cat *.out | grep HEDGE_WAIST | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.hwaist

    rm -f sim_${n}_out.zip
    zip sim_${n}_out.zip *.out inputgen/*.csv
    rm -f *.out
    rm -f inputgen/*.csv
done

set +x

echo "#### RESULTS ####"
echo ""

for n in ${NREPLICAS//,/ }
do
    echo "REPLICAS: ${n}"
    echo -n "DURATION: "
    awk -v ORS=, '{ print $0 }' sim_${n}.dur
    echo ""
    echo -n "PCP: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pcp
    echo ""
    echo -n "PVN: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pvn
    echo ""
    echo -n "HEDGED: "
    awk -v ORS=, '{ print $0 }' sim_${n}.hedged
    echo ""
    echo -n "NUM_PROC_SUCC: "
    awk -v ORS=, '{ print $0 }' sim_${n}.succ
    echo ""
    echo -n "NUM_PROC_FAILED: "
    awk -v ORS=, '{ print $0 }' sim_${n}.fail
    echo ""
    echo -n "HEDGE_WAIST: "
    awk -v ORS=, '{ print $0 }' sim_${n}.hwaist
    echo ""
done
