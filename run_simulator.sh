#!/bin/bash

date
set -x

echo "ROUND_START: ${ROUND_START:=1}"
echo "ROUND_END: ${ROUND_END:=4}"
echo "INSTANCES: ${INSTANCES:=1 2 4}"
echo "DURATION: ${DURATION:=300}"
echo "LOAD_PER_INSTANCE: ${LOAD_PER_INSTANCE:=40}"
echo "RESULTS_PATH: ${RESULTS_PATH:=/tmp/simulation/}"
mkdir -p $RESULTS_PATH

echo "PREFIX_RESULTS_NAME: ${PREFIX_RESULTS_NAME}"
echo "DATA_PATH: ${DATA_PATH}"
echo "INPUT_FILE_NAME: ${INPUT_FILE_NAME}"
echo "NUMBER_OF_INPUT_FILES: ${NUMBER_OF_INPUT_FILES}"
echo "DEBBUG_LOG: ${DEBBUG_LOG:='false'}"

for instance in ${INSTANCES};
do
	for round in `seq ${ROUND_START} ${ROUND_END}`
	do
    	echo ""
		echo "round ${round}: Simulating ${instance} instance(s)..."
	    NUMBER_OF_SERVERS=$instance DURATION=$DURATION LOAD=$(($LOAD_PER_INSTANCE * $instance)) RESULTS_PATH=$RESULTS_PATH DATA_PATH=$DATA_PATH INPUT_FILE_NAME=$INPUT_FILE_NAME NUMBER_OF_INPUT_FILES=$NUMBER_OF_INPUT_FILES RESULTS_NAME="${PREFIX_RESULTS_NAME}_${DURATION}_${LOAD}_${instance}_${round}" DEBBUG_LOG=$DEBBUG_LOG python3 __main__.py
		echo "round ${round}: Finished."
		echo ""
	done
done

