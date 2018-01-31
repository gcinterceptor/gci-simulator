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
            
def _initiate_csv_files(results_path, servers_number, load, availability_rate, communication_rate):
    _initiate_request_csv_files(results_path, servers_number, load, availability_rate, communication_rate)

def _initiate_request_csv_files(results_path, servers_number, load, availability_rate, communication_rate):
    request_data = [["id", "sent_time", "arrived_time", "finished_time", "latency_time", "server_id", "redirects", "refused"]]
    rsc_file_name = "request_" + str(servers_number) + "_" + load + "_" + str(availability_rate) + "_" + str(communication_rate) + ".csv"
    csv_writer(request_data, results_path + "/" + rsc_file_name)

