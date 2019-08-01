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
    