from log import csv_writer
import math

def log_request(requests, results_path, servers_number, scenario, load):
    data = list()

    for request in requests:
        data.append([request.id, request.created_at, request._sent_time, request._arrived_time,
                     request._finished_time, request._latency_time, request._interrupted_time, 
                     request.redirects, request.server_id])

    file_path = results_path + "/request_" + str(servers_number) + "_" + scenario + "_const_" + load + ".csv"
    csv_writer(data, file_path)
    
    