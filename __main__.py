from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from metrics import log_latency, log_time_in_server, log_server_metrics
from log import iniciate_csv_files
from config import get_config
import simpy, os, sys, time

def create_directory(path):
    if not os.path.isdir(path):
        os.mkdir(path)

def main():
    before = time.time()

    args = sys.argv
    SIM_DURATION_SECONDS = float(args[1])
    scenario = args[2]
    load = args[3]
    if len(args) == 5:
        results_path = args[4]
    else:
        results_path = "results"

    create_directory(results_path)
    iniciate_csv_files(results_path, scenario, load)

    if scenario == 'control':
        gc_conf = get_config('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
        gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')

    if load == 'high':
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-150 max_requests-inf')
    elif load == 'low':
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-35 max_requests-inf')

    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.00001')
    requests_conf = get_config('config/request.ini', 'request service_time-0.0035 memory-0.001606664')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

    env = simpy.Environment()
    server_control = ServerWithGCI(env, 1, server_conf, gc_conf, gci_conf)
    load_balancer_control = LoadBalancer(env, server_control, loadbalancer_conf)
    clients_control = Clients(env, load_balancer_control, client_conf, requests_conf)

    for time_stamp in range(1, int(SIM_DURATION_SECONDS + 1)):
        env.run(until=time_stamp)
        log_server_metrics(time_stamp, results_path, server_control, scenario, load)
        log_latency(time_stamp, results_path, clients_control.requests, scenario, load)
        log_time_in_server(time_stamp, results_path, clients_control.requests, scenario, load)

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()