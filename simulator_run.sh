#!/bin/bash

SIM_DURATION_SECONDS=$1
SCENARIO=$2
LOAD=$3
RESULTS_PATH=$4

echo "Simulating..."
MAIN_PATH=$(pwd)/__main__.py
python3 __main__.py $SIM_DURATION_SECONDS $SCENARIO $LOAD $RESULTS_PATH
