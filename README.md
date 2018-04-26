GCI-Simulator
===
> The **G**arbage **C**ollector **C**ontrol **I**nterceptor (**[GCI](https://github.com/gcinterceptor/gci-go)**) **Simulator** 
is the simulator used to study GCI on web cloud applications where multiples instances 
of services are used. The main objective here is to simulate web applications using GCI and not using GCI.

### Frameworks & Libraries 
* [SimPy](https://simpy.readthedocs.io/en/latest/)

### Dependencies
* [Python3 interpreter](https://www.python.org/downloads/)
* [SimPy](https://simpy.readthedocs.io/en/latest/simpy_intro/installation.html)

### Requirements
* Have experimental results CSV’s with the following information: 
  * **Service time**: A CSV with the service time of each request processed in experiment. 
  * **Processed requests** and **shedded request** at the same window: A **CSV file** with the number of processed requests until start shedding and the number of shedded requests at the this shedding. It is a requirement to simulate GCI on experiments only, there is no need of it to simulate GCI off experiments.

### How to run
#### Parameters
* **number_of_servers**: The number of servers to be used.
* **duration**: How much time the simulation should take in seconds (must be integer or float).
* **scenario**: There is two scenarios, control and baseline. Control means servers using GCI and baseline means servers with no GCI.
* **load**: An integer meaning how much requests the load balance must to distribute to the servers.
* **results_path**: The path where the simulator should put the results.
* **results_name**: A name to simulator name the results.
* **setup_info_name**: A name to simulator name the setup information files.
* **data_path**: The path where the CSV files of experimental results are.
* **service_time_file_name**: The name of a file with the service time of each request processed in an experiment.
* **service_time_data_column**: The column number of that files (integer).
* **round_start**: It defines the ID to identify the first simulation result file.
* **round_end**: It defines the ID to identify the last simulation result file. It also means the number of simulations to be executed.
* **shedding_file_name**: The name of the CSV file containing a pair of values needed to do shedding. 
* **shedding_number_of_files**: The number of shedding files.

#### Execution
After have cloned the simulator, move to the right director and execute one of those commands below. The command at Baseline simulates an experiment with **no GCI** on Servers, at control simulates with servers **using GCI**. The parameters must be passed as environment variables. Follow below a full example.

* ##### **Baseline**
  * NUMBER_OF_SERVERS="number_of_server" DURATION="duration" SCENARIO="**baseline**" LOAD="load" RESULTS_PATH="results_path" RESULTS_NAME="results_name" SETUP_INFO_NAME="setup_info_name" DATA_PATH="data_path" SERVICE_TIME_FILE_NAME="service_time_file_name" SERVICE_TIME_DATA_COLUMN="service_time_data_column" ROUND_START="round_start" ROUND_END="round_end" bash run_simulator.sh  
* ##### **Control**
  * NUMBER_OF_SERVERS="number_of_server" DURATION="duration" SCENARIO="**control**" LOAD="load" RESULTS_PATH="results_path" RESULTS_NAME="results_name" SETUP_INFO_NAME="setup_info_name" DATA_PATH="data_path" SERVICE_TIME_FILE_NAME="service_time_file_name" SERVICE_TIME_DATA_COLUMN="service_time_data_column" ROUND_START="round_start" ROUND_END="round_end" **SHEDDING_FILE_NAME="shedding_file_name" SHEDDING_NUMBER_OF_FILES="shedding_number_of_files"** bash run_simulator.sh  

Please, pay attention that the script run_simulation.sh already has some of these parameters with default values that make easier run simulations. 

### Results
The simulation will generate two kind of files: an information file and a CSV file. Both of files may be named as said before. The information file keeps the parameters used to simulate and the CSV file keeps data of each request processed during the simulation. The column description follow below. 
* **id**: The request identify.
* **created_time**: The moment at the request was created.
* **latency**: The all life time of a request processed startind at its created moment until it be processed and finished.
* **service_time**: Is only the time that the request have passed in a server. 
* **done**: It gives the request state. True if the request was processed and False otherwise. 
* **times_forwarded**: It gives how many times a request was sent to a server.
