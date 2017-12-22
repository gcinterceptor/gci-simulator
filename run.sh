#!/bin/bash

REPETIONS_NUMBER=10
SIMULATION_TIME=900
SERVERS_NUMBER="1 2 3 4 5 6 7 8 9 10 25 50 100"

LOAD="low
high"

SCENARIO="control
baseline"

AVAILABILITY_RATE="0.5 1 2 3 4 5 6 7 8 9 10 50 100"

#AVG REQUEST SERVICE TIME BY AVG COMPONENTS COMMUNICATION TIME
Y_PARAMETER="0.5 1 2 3 4 5 6 7 8 9 10 50 100"

mkdir results 2> /dev/null

for RN in $(seq $REPETIONS_NUMBER)
do
	for SN in $SERVERS_NUMBER
	do
		for SCENE in $SCENARIO
		do
			for LD in $LOAD
			do
				for AR in $AVAILABILITY_RATE
				do
					mkdir results/$RN 2> /dev/null
					echo "Servers Number=$SN, Simulation Time=$SIMULATION_TIME, Scenario=$SCENE, Load=$LD, Availability Rate=$AR"
					python3 __main__.py $SN $SIMULATION_TIME $SCENE $LD $AR results/$RN &
				done
			done
		done
	done 
done
