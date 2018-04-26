#!/bin/bash

date
set -x

echo "ROUND_START: ${ROUND_START:=1}"
echo "ROUND_END: ${ROUND_END:=1}"
echo "NUMBER_OF_SERVERS: ${NUMBER_OF_SERVERS:=1}"
echo "DURATION: ${DURATION:=120}"
echo "LOAD: ${LOAD:=$((${NUMBER_OF_SERVERS} * 80))}"
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

for round in `seq ${ROUND_START} ${ROUND_END}`
do
    echo ""
    echo "round ${round}: Simulating ${NUMBER_OF_SERVERS} instance(s)..."
	python3 __main__.py $round
	echo "round ${round}: Finished."
    echo ""
done 

INFO="number of service instances: ${NUMBER_OF_SERVERS}\nsimulation time duration: ${DURATION}\nconfiguration scenario: ${SCENARIO}\nworkload per instance: $(($LOAD / $NUMBER_OF_SERVERS))req/sec\ngeneral workload: ${LOAD}req/sec"
echo -e $INFO > $RESULTS_PATH$SETUP_INFO