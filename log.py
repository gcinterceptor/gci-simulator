import csv


def csv_writer(data, path, mode):
    with open(path, mode, newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)


def log_request(requests, results_path, results_name, mode):
    data = []
    if mode == "w":
        data.append(["timestamp", "status", "request_time"])

    for request in requests:
        data.append([request.finished_time * 1000, request.status, request._latency * 1000])

    file_path = results_path + "/" + results_name + ".csv"
    csv_writer(data, file_path, mode)


def log_debbug(requests, debbug_path, debbug_name, mode):
    data = []
    if mode == "w":
        data.append(["id", "created_time", "finished_time", "latency",
                     "service_time", "status", "done", "times_forwarded"])

    for request in requests:
        data.append([request.id, request.created_time, request.finished_time, request._latency,
                     request.service_time, request.status, request.done, request.times_forwarded])

    file_path = debbug_path + "/" + debbug_name + "_debbug.csv"
    csv_writer(data, file_path, mode)

