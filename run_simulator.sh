#!/bin/bash

date
set -x

echo "ROUND_START: ${ROUND_START:=1}"
echo "ROUND_END: ${ROUND_END:=1}"
echo "INSTANCES: ${ISNTANCES:=1 2 4}"
echo "DURATION: ${DURATION:=120}"
echo "LOAD_PER_INSTANCE: ${LOAD_PER_INSTANCE:=80}"
echo "SCENARIO: ${SCENARIO:=control}"
echo "RESULTS_PATH: ${RESULTS_PATH:=/tmp/simulation/${NUMBER_OF_SERVERS}i/${SCENARIO}/}"
mkdir -p $RESULTS_PATH

echo "RESULTS_NAME: ${RESULTS_NAME}"
echo "DATA_PATH: ${DATA_PATH}"
echo "SERVICE_TIME_FILE_NAME: ${SERVICE_TIME_FILE_NAME}"
echo "SERVICE_TIME_DATA_COLUMN: ${SERVICE_TIME_DATA_COLUMN}"
echo "SHEDDING_FILE_NAME: ${SHEDDING_FILE_NAME}"
echo "SHEDDING_NUMBER_OF_FILES: ${SHEDDING_NUMBER_OF_FILES}"
echo "SETUP_INFO: ${SETUP_INFO:=setup_info.txt}"


for instance in ${INSTANCES};
do
	for round in `seq ${ROUND_START} ${ROUND_END}`
	do
		echo ""
		echo "round ${round}: Simulating ${instance} instance(s)..."
		LOAD=$(($LOAD_PER_INSTANCE * $instance)) NUMBER_OF_SERVERS=$instance python3 __main__.py
		mkdir -p $RESULTS_PATH${instance}i/
		mv "${RESULTS_PATH}${RESULTS_NAME}.csv" "${RESULTS_PATH}${instance}i/${RESULTS_NAME}_${round}.csv"
		echo "round ${round}: Finished."
		echo ""
	done 
	
	INFO="number of service instances: ${instance}\nsimulation time duration: ${DURATION}\nconfiguration scenario: ${SCENARIO}\nworkload per instance: ${LOAD_PER_INSTANCE}req/sec\ngeneral workload: $(($LOAD_PER_INSTANCE * ${instance}))req/sec"
	echo -e $INFO > $RESULTS_PATH${instance}i/$SETUP_INFO
	
done

