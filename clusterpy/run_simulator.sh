#!/bin/bash

BLUE='\033[0;34m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

date
set -x

echo -e "${GREEN}STARTING SIMULATION${NC}"

echo "ROUND_START: ${ROUND_START:=1}"
echo "ROUND_END: ${ROUND_END:=4}"
echo "INSTANCES: ${INSTANCES:=1 2 4}"
echo "DURATION: ${DURATION:=300}"
echo "LOAD_PER_INSTANCE: ${LOAD_PER_INSTANCE:=40}"
echo "RESULTS_PATH: ${RESULTS_PATH:=/tmp/simulation/}"
mkdir -p $RESULTS_PATH

echo "PREFIX_RESULTS_NAME: ${PREFIX_RESULTS_NAME}"
echo "DATA_PATH: ${DATA_PATH}"
echo -e "${YELLOW}INPUT_FILE_NAMES: ${INPUT_FILE_NAMES}${NC}"

for instance in ${INSTANCES};
do
	echo -e "${GREEN}SIMULATING ${instance} INSTANCE(S)...${NC}"
	for round in `seq ${ROUND_START} ${ROUND_END}`
	do
		echo ""
		echo -e "${BLUE}ROUND ${round}!${NC}"
		NUMBER_OF_SERVERS=$instance DURATION=$DURATION LOAD=$(($LOAD_PER_INSTANCE * $instance)) DATA_PATH=$DATA_PATH INPUT_FILE_NAMES=$INPUT_FILE_NAMES RESULTS_PATH=$RESULTS_PATH RESULTS_NAME="${PREFIX_RESULTS_NAME}_${DURATION}_${LOAD}_${instance}_${round}" python3 __main__.py
		echo "round ${round}: Finished."
		echo ""
	done
done

echo -e "${RED}SIMULATION FINISHED${NC}"
