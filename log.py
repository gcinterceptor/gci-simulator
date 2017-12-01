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

def iniciate_csv_files(results_path, scenario, load):
    if scenario == "control":
        _iniciate_control_csv_files(results_path, load)
    elif scenario == "baseline":
        _iniciate_baseline_csv_files(results_path, load)

def _iniciate_control_csv_files(results_path, load):
    server_data = [["heap_level", "remaining_requests", "gci_exe", "gc_exe", "gc_exe_time_sum", "processed_requests"]]
    latency_data = [["first_request", "last_request", "average_latency", "median", "p90", "p95", "p99", "p999"]]
    _iniciate_csv_files(latency_data, server_data, results_path, "control", load)

def _iniciate_baseline_csv_files(results_path, load):
    server_data = [["heap_level", "remaining_requests", "gc_exe", "gc_exe_time_sum", "processed_requests"]]
    latency_data = [["first_request", "last_request", "average_latency", "median", "p90", "p95", "p99", "p999"]]
    _iniciate_csv_files(latency_data, server_data, results_path, "baseline", load)

def _iniciate_csv_files(latency_data, server_data, results_path, scenario, load):
    rlc_file_name = "requests_latency_" + scenario + "_const_" + load + ".csv"
    ssc_file_name = "server_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(latency_data, results_path + "/" + rlc_file_name)
    csv_writer(server_data, results_path + "/" + ssc_file_name)