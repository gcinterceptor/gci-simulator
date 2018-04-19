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
	
echo "DATA_PATH: ${DATA_PATH}"
echo "SERVICE_TIME_FILE_NAME: ${SERVICE_TIME_FILE_NAME}"
echo "SERVICE_TIME_DATA_COLUMN: ${SERVICE_TIME_DATA_COLUMN}"
echo "SHEDDING_FILE_NAME: ${SHEDDING_FILE_NAME}"
echo "SHEDDING_NUMBER_OF_FILES: ${SHEDDING_NUMBER_OF_FILES}"

for round in `seq ${ROUND_START} ${ROUND_END}`
do
    echo ""
    echo "round ${round}: Simulating..."
	
	MAIN_PATH=$(pwd)/__main__.py
	python3 __main__.py $NUMBER_OF_SERVERS $DURATION $SCENARIO $LOAD $RESULTS_PATH $DATA_PATH $SERVICE_TIME_FILE_NAME $SERVICE_TIME_DATA_COLUMN $round $SHEDDING_FILE_NAME $SHEDDING_NUMBER_OF_FILES
	
	echo "round ${round}: Finished."
    echo ""
done 
