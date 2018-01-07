from log import csv_writer
import math

def log_request(requests, results_path, servers_number, scenario, load, availability_rate, communication_rate):
    data = list()

    for request in requests:
        data.append([request.id, request._sent_time, request._arrived_time,
                     request._finished_time, request._latency_time, request.server_id])

    file_path = results_path + "/request_" + str(servers_number) + "_" + scenario + "_" + load + "_" + str(availability_rate) + "_" + str(communication_rate) + ".csv"
    csv_writer(data, file_path)
    
    
