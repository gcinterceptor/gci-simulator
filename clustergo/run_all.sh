#!/bin/bash

set -x
set -e

rm -f sim_* *.out

###### 1X

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.003 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# head -n1 r2_1.out > sim_sample.csv
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE BASELINE 1x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.003 HT=847.3845 HEDGE_CANCELLATION=false NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# head -n1 r2_1.out > sim_sample.csv
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE 1x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.003 HT=847.3845 HEDGE_CANCELLATION=true NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# head -n1 r2_1.out > sim_sample.csv
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE CANC 1x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=true LITTLE_OMEGA=0.0001 BIG_OMEGA=0.003 HT=-1 HEDGE_CANCELLATION=false NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# head -n1 r2_1.out > sim_sample.csv
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE CTC 1x'

# # ###### 10X

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_10x/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.03 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_10x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_10x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# head -n1 r2_1.out > sim_sample.csv
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE BASELINE 10x'

HT=1047.45

rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_10x/*
cd ~/old/fireman/repos/gci-simulator/clustergo
ENABLE_CCT=false BIG_OMEGA=0.03 HT=${HT} NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_10x/
cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_10x/
for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
head -n1 r2_1.out > sim_sample.csv
cat sim.out | shuf -n 1000000 >> sim_sample.csv
aplay /usr/share/sounds/speech-dispatcher/test.wav
notify-send 'DONE HEDGE 10x'

rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_10x/*
cd ~/old/fireman/repos/gci-simulator/clustergo
ENABLE_CCT=false BIG_OMEGA=0.03 HT=${HT} HEDGE_CANCELLATION=true NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_10x/
cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_10x/
for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
head -n1 r2_1.out > sim_sample.csv
cat sim.out | shuf -n 1000000 >> sim_sample.csv
aplay /usr/share/sounds/speech-dispatcher/test.wav
notify-send 'DONE HEDGE CANC 10x'

rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_10x/*
cd ~/old/fireman/repos/gci-simulator/clustergo
ENABLE_CCT=true LITTLE_OMEGA=0.0001 BIG_OMEGA=0.03 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_10x/
cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_10x/
for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
head -n1 r2_1.out > sim_sample.csv
cat sim.out | shuf -n 1000000 >> sim_sample.csv
aplay /usr/share/sounds/speech-dispatcher/test.wav
notify-send 'DONE CTC 10x'

# # ###### 100X

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_100x/
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.3 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE BASELINE 100x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_100x/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.3 HT=847.3845 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE 100x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_100x/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=false BIG_OMEGA=0.3 HT=847.3845 HEDGE_CANCELLATION=true NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_canc_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE CANC 100x'

# rm -rf /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_100x/*
# cd ~/old/fireman/repos/gci-simulator/clustergo
# ENABLE_CCT=true LITTLE_OMEGA=0.0001 BIG_OMEGA=0.3 HT=-1  NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE CTC 100x'