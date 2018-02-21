import csv


def csv_writer(data, path):
    with open(path, "w", newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)


def log_request(requests, results_path, results_name, scenario, load):
    data = [["id", "created_time", "latency", "service_time",  "done", "times_forwarded"]]

    for request in requests:
        data.append([request.id, request.created_time, request._latency, request.service_time, request.done, request.times_forwarded])

    file_path = results_path + "/" + results_name + "_" + scenario + "_const_" + load + ".csv"
    csv_writer(data, file_path)


def txt_writer(file_name, text):
    arq = open(file_name, 'w')
    arq.write(text)
    arq.close()