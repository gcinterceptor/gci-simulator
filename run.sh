#!/bin/bash

REPETIONS_NUMBER=10
SIMULATION_TIME=600
SERVERS_NUMBER="1 2 3 4 5 6 7 8 9 10 50 100"

LOAD="low
high"

SCENARIO="control
baseline"

AVAILABILITY_RATE="0.5 1 2 3 4 5 6 7 8 9 10 50"

#AVG COMPONENTS COMMUNICATION TIME BY AVG UNAVAILABLE TIME
COMMUNICATION_RATE="0.025"

mkdir results-1 2> /dev/null

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
					timestamp=$(date +%s)
					for SCENE in $SCENARIO
					do
						mkdir results-1/$RN 2> /dev/null
						echo "REP=$RN, Servers Number=$SN, Simulation Time=$SIMULATION_TIME, Scenario=$SCENE, Load=$LD, Availability Rate=$AR, Communication_Rate=$CR"
						python3 __main__.py $SN $SIMULATION_TIME $SCENE $LD $AR $SR $CR results-1/$RN $timestamp
					done
				done
			done
		done
	done
done
