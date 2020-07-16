#!/bin/bash

set -x

###### 1X

# ENABLE_CCT=false BIG_OMEGA=0.003 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE BASELINE 1x'

# ENABLE_CCT=false BIG_OMEGA=0.003 HT=847.3845 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE 1x'

# ENABLE_CCT=true BIG_OMEGA=0.003 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE CTC 1x'

# ###### 10X

# ENABLE_CCT=false BIG_OMEGA=0.03 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_10x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_10x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE BASELINE 10x'

# ENABLE_CCT=false BIG_OMEGA=0.03 HT=1046.561 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_10x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_10x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE 10x'

# ENABLE_CCT=true BIG_OMEGA=0.03 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_10x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_10x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE CTC 10x'

# ###### 100X

ENABLE_CCT=false BIG_OMEGA=0.3 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_100x/
cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/baseline_100x/
for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
cat sim.out | shuf -n 1000000 >> sim_sample.csv
cd -
aplay /usr/share/sounds/speech-dispatcher/test.wav
notify-send 'DONE BASELINE 100x'

# ENABLE_CCT=false BIG_OMEGA=0.3 HT=847.3845 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/hedge_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE HEDGE 100x'

# ENABLE_CCT=true BIG_OMEGA=0.3 HT=-1 NREPLICAS=2,4,8,16,32,64 ./run_exp.sh
# mv sim_* /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_100x/
# cd /home/danielfireman/tese/resultados/data_cmp_sim_hedge/ctc_100x/
# for i in 2 4 8 16 32 64; do unzip -o sim_${i}_out.zip; done
# cat r*.out | grep -h -w "^[[:digit:]].*\,[503\,|200\,].*" > sim.out
# cat sim.out | shuf -n 1000000 >> sim_sample.csv
# cd -
# aplay /usr/share/sounds/speech-dispatcher/test.wav
# notify-send 'DONE CTC 100x'