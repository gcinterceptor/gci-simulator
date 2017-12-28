from log import csv_writer
import math

def percentile(l, func, p):
    """Returns the value within the list l representing pth percentile."""
    pos = math.ceil((p / 100) * len(l))
    return func(l[pos - 1])

def get_average_median_and_percentiles(requests, func):
    ordered_requests = sorted(requests, key=func)

    average = sum(func(request) for request in requests) / len(requests)
    median = percentile(ordered_requests, func, 50)
    p90 = percentile(ordered_requests, func, 90)
    p95 = percentile(ordered_requests, func, 95)
    p99 = percentile(ordered_requests, func, 99)
    p999 = percentile(ordered_requests, func, 99.9)

    return [average, median, p90, p95, p99, p999]

def log_request_metrics(time_stamp, results_path, metric_name, requests, scenario, load, func):
    file_path = results_path + "/" + metric_name + "_" + scenario + "_const_" + load + ".csv"
    results = [[time_stamp] + get_average_median_and_percentiles(requests, func)]
    csv_writer(results, file_path)

def log_latency(time_stamp, results_path, requests, scenario, load):
    log_request_metrics(time_stamp, results_path, "requests_latency_metrics", requests, scenario, load, lambda x: x._latency_time)

def log_time_in_server(time_stamp, results_path, requests, scenario, load):
    log_request_metrics(time_stamp, results_path, "time_in_server_metrics", requests, scenario, load, lambda x: x._time_in_server)

def log_server_metrics(time_stamp, results_path, server, scenario, load):
    heap_level = server.heap.level
    requests_in_queue = len(server.queue.items)
    gc_exe = server.gc.times_performed
    gc_exe_sum = server.gc.gc_exec_time_sum
    processed_requests = server.processed_requests
    times_interrupted = server.times_interrupted

    results = [[time_stamp, heap_level, requests_in_queue, gc_exe, gc_exe_sum, processed_requests, times_interrupted]]
    if (scenario == "control"):
        gci_exe = server.gci.times_performed
        results[0].insert(3, gci_exe)

    file_path = results_path+ "/server_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(results, file_path)

def log_request(requests, results_path, scenario, load):
    data = list()

    for request in requests:
        data.append([request.id, request._time_in_queue, request._time_in_server,
                     request.created_at, request._arrived_time, request._attended_time,
                     request._finished_time, request._latency_time])

    file_path = results_path + "/request_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(data, file_path)
