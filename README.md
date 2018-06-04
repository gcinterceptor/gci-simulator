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
* Have an experimental results CSV file with the following information: 
  * **Status**: The http status that each request has received.
  * **Request time**: the service time of each request processed in experiment. 
  
### How to run
#### Parameters
* **instances**: The numbers of servers to be used. It isn't straightforward, but you need pass a string with each number os instances to be simulated. Example: INSTANCES="1 2 4".
* **duration**: How much time the simulation should take in seconds (must be integer or float).
* **load_per_instance**: An integer meaning how much requests the load balance must to distribute to each server.
* **results_path**: The path where the simulator should put the results.
* **prefix_results_name**: A pattern to simulator use it as a prefix in result names.
* **data_path**: The path where the CSV file of the experimental results is kept.
* **input_file_name**: The name of the CSV file with the service time and status http of each request processed in an experiment.
* **round_start**: It defines the ID to identify the first simulation result file.
* **round_end**: It defines the ID to identify the last simulation result file. It also means the number of simulations to be executed.
* **debbug_log**: It is used to generate more detailed results.

#### Execution
After have cloned the simulator, move to the right directory and execute the command below. Note that the input file will shape your simulation behavior.

  * **INSTANCES**="instances" **DURATION**="duration" **LOAD_PER_INSTANCE**="load_per_instance" **RESULTS_PATH**="results_path" **PREFIX_RESULTS_NAME**="prefix_results_name" **DATA_PATH**="data_path" **INPUT_FILE_NAME**="input_file_name" **ROUND_START**="round_start" **ROUND_END**="round_end" **bash** **run_simulator.sh**  

Please, pay attention that the script run_simulation.sh already has some of these parameters with default values that make easier run simulations. 

### Results
The simulation will generate one or two kind of result files: If you pass no debbug key the
simulate will generate only a simple file result containing three columns. If you pass DEBBUG=true, a more detailed 
result will be created. The debbug result has the same name as the normal file but with debbug 
as sufix. The bullets below explains what each column represent.
* **id**: A number that identifies some request.
* **timestamp**: The moment at the request was finished.
* **created_time**: The moment at the request was created.
* **latency**: The all life time of a request processed starting at its created moment until it be processed and finished.
* **service_time**: It is only the time that the request have passed in a server. 
* **status**: It gives the http status of a request. 200 means successfully request, 503 means failed requests.  
* **done**: It gives the request state. True if the request was processed and False otherwise. 
* **times_forwarded**: It gives how many times a request was sent to a server.
* **hops**: Each service time that a request has gain in some server.

Since who run a simulation have chose a prefix name, the result name radical is always named 
based on simulation configuration following the pattern "duration_load_instances_round", where "duration" means the
simulation duration, "load" means how much request where sent each second, "instances" means how much instances where used and
"round" means which round each file belongs to.
