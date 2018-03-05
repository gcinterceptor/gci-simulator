GCI-Simulator
===
> The **G**arbage **C**ollector **C**ontrol **I**nterceptor (**[GCI]()**) **Simulator** 
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
  * **Processed requests** and **shedded request** at the same window: A CSV with the number of processed requests until start shedding and the number of shedded requests at the this shedding.

### How to run
#### Parameters
* **number_of_server**: The number of servers to be used.
* **sim_duration**: How much time the simulation should take in seconds (must be integer or float).
* **scenario**: There is two scenarios, control and baseline. Control means simulation with servers using GCI and baseline means exactly the opposite.
* **load**: An integer meaning how much requests the load balance must to distribute to the servers.
* **result_path**: The path where the simulator should keep the results.
* **data_path**: The path where the CSV files of experimental results are.
* **service_time_file_name**: The name of a file with the service time of each request processed in an experiment.
* **service_time_data_column**: The column number of that files (integer).
* **simulation_number**: An ID to identify each simulation result file.
* **shedding_file_name**: The name of the file containing a pair of values needed to do shedding. 
* **shedding_number_of_files**: The number of shedding files.

#### Execution
After have cloned the simulator, move to the right director and execute one of those commands below. The command at Baseline simulates an experiment with no GCI on Servers, at control simulates with servers using GCI.

* ##### Baseline
  * bash simulation_run.sh number_of_server sim_duration **baseline** load result_path data_path service_time_file_name service_time_data_column simulation_number 
* ##### Control
  * bash simulation_run.sh number_of_server sim_duration **control** load result_path data_path service_time_file_name service_time_data_column simulation_number **shedding_file_name shedding_number_of_files**


### Results

