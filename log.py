import csv


def csv_writer(data, path, mode):
    with open(path, mode, newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)


def log_request(requests, results_path, results_name, mode):
    data = []
    if mode == "w":
        data.append(["id", "created_time", "latency", "service_time",  "done", "times_forwarded"])

    for request in requests:
        data.append([request.id, request.created_time, request._latency, request.service_time, request.done, request.times_forwarded])

    file_path = results_path + "/" + results_name + ".csv"
    csv_writer(data, file_path, mode)
