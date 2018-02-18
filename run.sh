#!/bin/bash

OUTPUT_PATH=$1

REPETIONS_NUMBER=5
SIMULATION_TIME=600
SERVERS_NUMBER="2 4 8 16 32 64"

LOAD="low
high"

SHEDDED_REQUESTS_RATE="0.2 0.1 0.01 0.001 0.0001 0.00001"

#AVG COMPONENTS COMMUNICATION TIME BY AVG UNAVAILABLE TIME
NETWORK_COMMUNICATION_TIME="0.002"

REQUESTS_CPU_TIME="0.010"

mkdir $OUTPUT_PATH 2> /dev/null

for RN in $(seq $REPETIONS_NUMBER)
do
	for SN in $SERVERS_NUMBER
	do
		for LD in $LOAD
		do
			for SRR in $SHEDDED_REQUESTS_RATE
			do
				for NCT in $NETWORK_COMMUNICATION_TIME
				do
					for RCT in $REQUESTS_CPU_TIME
					do
						mkdir $OUTPUT_PATH/$RN 2> /dev/null
						echo "REP=$RN, Servers Number=$SN, Simulation Time=$SIMULATION_TIME, Load=$LD, Shedded Requests Rate=$SRR, Network Communication Time=$NCT, Requests CPU Time=$RCT"
						python3 __main__.py $SN $SIMULATION_TIME $LD $SRR $NCT $RCT $OUTPUT_PATH/$RN
					done
				done
			done
		done
	done
done
