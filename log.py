import csv


def csv_writer(data, path):
    with open(path, "a", newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)


def log_request(requests, results_path, scenario, load):
    data = [["id", "created_time", "latency"]]

    for request in requests:
        data.append([request.id, request.created_time, request._latency])

    file_path = results_path + "/request_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(data, file_path)


def log_gc(gc_count, gc_time, results_path, scenario, load):
    data = [["ts", "count", "time"]]

    for i in range(len(gc_count)):
        data.append([i + 1, gc_count[i], gc_time[i]])

    file_path = results_path + "/gc_status_" + scenario + "_const_" + load + ".csv"
    csv_writer(data, file_path)