#!/bin/bash

set -x

DURATIONMS=3600000
BIG_OMEGA=0.003
MU=0.0036
LITTLE_OMEGA=0.0001
NRUNS=${NRUNS:=50}  # number of repetitions
NREPLICAS=${NREPLICAS:=1} # comma-separated number of replicas

for n in ${NREPLICAS//,/ }
do
    for run in `seq 1 ${NRUNS}`
    do
        cd inputgen/;
        go run main.go --duration=${DURATIONMS} --r=$n --mu=${MU} --bigOmega=${BIG_OMEGA} --littleOmega=${LITTLE_OMEGA}
        cd ..
        input=""
        for i in `seq 1 ${n}`
        do
            input+="inputgen/input_${i}.csv,"
        done
        go run main.go --rate=1 --warmup=0 --d=$(( DURATIONMS * n ))ms --i=$input > r${n}_${run}.out
    done

    cat *.out | grep PCP | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pcp
    cat *.out | grep PVN | cut -d' ' -f1 | cut -d: -f2 > sim_${n}.pvn

    zip sim_${n}_out.zip r${n}_${run}.out
    rm *.out

    echo -e "\n ### REPLICA: PCP ###\n"
    awk -v ORS=, '{ print $0 }' sim_${n}.pcp
    echo -e "\n ###### \n"

    echo -e "\n ### REPLICA: PVN ###\n"
    awk -v ORS=, '{ print $0 }' sim_${n}.pvn
    echo -e "\n ###### \n"
done
