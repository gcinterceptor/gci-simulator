from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from log import iniciate_csv_files, csv_writer
from config import get_config
import simpy, os, time, math

def create_directory(logs_path, results_path):
    if not os.path.isdir(logs_path):
        os.mkdir(logs_path)

    if not os.path.isdir(results_path):
        os.mkdir(results_path)

def percentile(l, p):
    """Returns the value within the list l representing pth percentile."""
    s = sorted(l, key=lambda x: x._latency_time)
    pos = math.ceil((p / 100) * len(s))
    return s[pos - 1]._latency_time

def log_latency(results_path, requests, scenario, load):
    first_request = requests[0]._latency_time
    last_request = requests[-1]._latency_time
    average_latency = sum(request._latency_time for request in requests) / len(requests)

    ordered_requests = sorted(requests, key=lambda x: x._latency_time)
    median = percentile(ordered_requests, 50)
    p90 = percentile(ordered_requests, 90)
    p95 = percentile(ordered_requests, 95)
    p99 = percentile(ordered_requests, 99)
    p999 = percentile(ordered_requests, 99.9)

    file_path = results_path + "/requests_latency_" + scenario+ "_const_" + load + ".csv"
    results = [[first_request, last_request, average_latency, median, p90, p95, p99, p999]]
    csv_writer(results, file_path)

def log_server_data(results_path, server, scenario, load):
    heap_level = server.heap.level
    remaining_requests = len(server.queue.items)
    gci_exe = server.gci.times_performed
    gc_exe = server.gc.times_performed
    gc_exe_sum = server.gc.gc_exec_time_sum
    processed_requests = server.processed_requests

    file_path = results_path+ "/server_status_" + scenario + "_const_" + load + ".csv"
    results = [[heap_level, remaining_requests, gci_exe, gc_exe, gc_exe_sum, processed_requests]]
    csv_writer(results, file_path)

def main():
    before = time.time()

    log_path = 'logs'
    results_path = "results"
    create_directory(log_path, results_path)
    iniciate_csv_files(results_path, "control", "low")
    iniciate_csv_files(results_path, "baseline", "low")
    env = simpy.Environment()

    client_conf = get_config('config/clients.ini', 'clients sleep_time-0.00001 create_request_rate-35 max_requests-inf')
    gc_conf = get_config('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
    gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.0035')
    requests_conf = get_config('config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

    server_control = ServerWithGCI(env, 1, server_conf, gc_conf, gci_conf, log_path)
    load_balancer_control = LoadBalancer(env, server_control, loadbalancer_conf, log_path)
    clients_control = Clients(env, load_balancer_control, client_conf, requests_conf, log_path)

    server_baseline = Server(env, 1, server_conf, gc_conf, log_path)
    load_balancer_baseline = LoadBalancer(env, server_baseline, loadbalancer_conf, log_path)
    clients_baseline = Clients(env, load_balancer_baseline, client_conf, requests_conf, log_path)

    SIM_DURATION_SECONDS = 1
    env.run(until=SIM_DURATION_SECONDS)

    log_server_data(results_path, server_control, "control", "low")
    log_latency(results_path, clients_control.requests, "control", "low")
    log_latency(results_path, clients_baseline.requests, "baseline", "low")

    after = time.time()
    print("Time of execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()