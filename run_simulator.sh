#!/bin/bash

date
set -x

echo "ROUND_START: ${ROUND_START:=1}"
echo "ROUND_END: ${ROUND_END:=1}"
echo "INSTANCES: ${ISNTANCES:=1 2 4}"
echo "DURATION: ${DURATION:=300}"
echo "LOAD_PER_INSTANCE: ${LOAD_PER_INSTANCE:=40}"
echo "RESULTS_PATH: ${RESULTS_PATH:=/tmp/simulation/}"
mkdir -p $RESULTS_PATH

echo "GENERAL_RESULTS_NAME: ${GENERAL_RESULTS_NAME}"
echo "DATA_PATH: ${DATA_PATH}"
echo "SERVICE_TIME_FILE_NAME: ${SERVICE_TIME_FILE_NAME}"

for instance in ${INSTANCES};
do
	for round in `seq ${ROUND_START} ${ROUND_END}`
	do
		echo "\nround ${round}: Simulating ${instance} instance(s)..."
		NUMBER_OF_SERVERS=$instance LOAD=$(($LOAD_PER_INSTANCE * $instance)) RESULTS_NAME="${GENERAL_RESULTS_NAME}_${DURATION}_${LOAD}_${instance}_${round}" python3 __main__.py
		echo "round ${round}: Finished.\n"
	done
done

