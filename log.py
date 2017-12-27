import logging, csv

def get_logger(path_file, logger_name):
    handler = logging.FileHandler(path_file, mode='w')

    formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s')
    handler.setFormatter(formatter)

    logger = logging.getLogger(logger_name)
    logger.setLevel(logging.DEBUG)
    logger.addHandler(handler)

    return logger

def csv_writer(data, path):
    with open(path, "a", newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)

def initiate_csv_files(results_path, scenario, load):
    _initiate_request_csv_files(results_path, scenario, load)
    _initiate_server_csv_files(results_path, scenario, load)

def _initiate_request_csv_files(results_path, scenario, load):
    latency_data = [["time_stamp", "average_latency", "median", "p90", "p95", "p99", "p999"]]
    rlc_file_name = "requests_latency_metrics_" + scenario + "_const_" + load + ".csv"
    csv_writer(latency_data, results_path + "/" + rlc_file_name)

    request_data = [["id", "time_in_queue", "time_in_server", "created_at", "arrived_time", "attended_time", "finished_time", "latency_time"]]
    rsc_file_name = "request_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(request_data, results_path + "/" + rsc_file_name)

    time_in_server_data = [["time_stamp", "average_time", "median", "p90", "p95", "p99", "p999"]]
    tic_file_name = "time_in_server_metrics_" + scenario + "_const_" + load + ".csv"
    csv_writer(time_in_server_data, results_path + "/" + tic_file_name)

def _initiate_server_csv_files(results_path, scenario, load):
    if scenario == "control":
        server_data = [["time_stamp", "heap_level", "requests_in_queue", "gci_exe", "gc_exe", "gc_exe_time_sum", "processed_requests", "times_interrupted"]]

    elif scenario == "baseline":
        server_data = [["time_stamp", "heap_level", "requests_in_queue", "gc_exe", "gc_exe_time_sum", "processed_requests", "times_interrupted"]]

    ssc_file_name = "server_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(server_data, results_path + "/" + ssc_file_name)
