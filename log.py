import csv


def csv_writer(data, path, mode):
    with open(path, mode, newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)


def log_request(requests, results_path, results_name, mode):
    data = []
    if mode == "w":
        data.append(["timestamp", "status", "latency", "hops"])

    for request in requests:
        data.append([request.finished_time * 1000, request.status, request._latency * 1000, " ".join(request.hops)])

    file_path = results_path + "/" + results_name + ".csv"
    csv_writer(data, file_path, mode)


def log_debbug(requests, debbug_path, debbug_name, mode):
    data = []
    if mode == "w":
        data.append(["id", "timestamp", "created_time", "latency",
                     "service_time", "status", "times_forwarded", "hops"])

    for request in requests:
        data.append([request.id, request.finished_time * 1000, request.created_time, request._latency,
                     request.service_time, request.status, request.times_forwarded, " ".join(request.hops)])

    file_path = debbug_path + "/" + debbug_name + "_debbug.csv"
    csv_writer(data, file_path, mode)

