#!/bin/bash

for n in ${NREPLICAS//,/ }
do
    echo ""
    echo "REPLICAS: ${n}"
    echo -n "PCP: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pcp
    echo ""; echo -n "PVN: "
    awk -v ORS=, '{ print $0 }' sim_${n}.pvn
    echo "";
    echo -n "THROUGHPUT: "
    awk -v ORS=, '{ print $0 }' sim_${n}.tp
    echo ""
    echo -n "HEDGE: "
    awk -v ORS=, '{ print $0 }' sim_${n}.hedged
    echo ""
done
