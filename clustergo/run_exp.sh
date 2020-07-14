#!/bin/bash

set -x

LITTLE_OMEGA=${LITTLE_OMEGA:=0.0001}
BIG_OMEGA=${BIG_OMEGA:=0.003}
MU=${MU:=0.0036}
DURATIONMS=${DURATIONMS:=1200000}
NRUNS=${NRUNS:=30}  # number of repetitions
NREPLICAS=${NREPLICAS:=1} # comma-separated number of replicas
HT=${HT:=-1} # Hedging threshold (-1 means no hedging)

for n in ${NREPLICAS//,/ }
do
    for run in `seq 1 ${NRUNS}`
    do
        echo -e "\n\n## Starting run $n : ${run} ##\n\n"
        cd inputgen/;
        go run main.go --duration=${DURATIONMS} --r=$n --mu=${MU} --bigOmega=${BIG_OMEGA} --littleOmega=${LITTLE_OMEGA}
        cd ..
        input=""
        for i in `seq 1 ${n}`
        do
            input+="inputgen/input_${i}.csv,"
        done
        go run main.go --rate=1 --warmup=0 --d=$(( DURATIONMS ))ms --ht=${HT} --i=$input > r${n}_${run}.out
        echo -e "\n\n## Finished run $n : ${run} ##\n\n"
    done
    cat *.out | grep PCP | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pcp
    cat *.out | grep PVN | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pvn
    cat *.out | grep THROUGHPUT | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.tp
    cat *.out | grep HEDGED | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.hedged

    rm sim_${n}_out.zip
    zip sim_${n}_out.zip *.out inputgen/*.csv
    rm *.out
    rm inputgen/*.csv
done

set +x

echo "#### RESULTS ####"
echo ""

for n in ${NREPLICAS//,/ }
do
    echo "REPLICAS: ${n}"
    echo -n "PCP: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pcp
    echo ""
    echo -n "PVN: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pvn
    echo ""
    echo -n "THROUGHPUT: "
    awk -v ORS=, '{ print $0 }' sim_${n}.tp
    echo ""
    echo -n "HEDGED: "
    awk -v ORS=, '{ print $0 }' sim_${n}.hedged
    echo ""
done
