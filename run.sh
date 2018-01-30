#!/bin/bash

OUTPUT_PATH=$1

REPETIONS_NUMBER=10
SIMULATION_TIME=600
SERVERS_NUMBER="2 4 8 50"

LOAD="low
high"

AVAILABILITY_RATE="0.1 0.5 1 2 10"

#AVG COMPONENTS COMMUNICATION TIME BY AVG UNAVAILABLE TIME
COMMUNICATION_RATE="0.025"

mkdir $OUTPUT_PATH 2> /dev/null

for RN in $(seq $REPETIONS_NUMBER)
do
	for SN in $SERVERS_NUMBER
	do
		for LD in $LOAD
		do
			for AR in $AVAILABILITY_RATE
			do
				for CR in $COMMUNICATION_RATE
				do
					mkdir $OUTPUT_PATH/$RN 2> /dev/null
					echo "REP=$RN, Servers Number=$SN, Simulation Time=$SIMULATION_TIME, Load=$LD, Availability Rate=$AR, Communication_Rate=$CR"
					python3 __main__.py $SN $SIMULATION_TIME $LD $AR $CR $OUTPUT_PATH/$RN
				done
			done
		done
	done
done
