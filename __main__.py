from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from metrics import log_latency, log_time_in_server, log_server_metrics, log_request
from log import initiate_csv_files
from config import get_config
import simpy, os, sys, time

def create_directory(path):
    if not os.path.isdir(path):
        os.mkdir(path)

def main():
    before = time.time()

    args = sys.argv

    NUMBER_OF_SERVERS = int(args[1])
    SIM_DURATION_SECONDS = float(args[2])
    scenario = args[3]
    load = args[4]

    if len(args) >= 6:
        results_path = args[5]
    else:
        results_path = "results"
    create_directory(results_path)

    if len(args) >= 7:
        log_path = args[6]
        create_directory(log_path)
    else:
        log_path = None

    initiate_csv_files(results_path, scenario, load)

    if load == 'high':
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-150 max_requests-inf')
        loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.006666667')

    elif load == 'low':
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-35 max_requests-inf')
        loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.028571429')

    else:
        raise Exception("INVALID LOAD")

    env = simpy.Environment()

    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

    servers = list()
    load_balancer = LoadBalancer(env, loadbalancer_conf, log_path)
    for i in range(NUMBER_OF_SERVERS):
        if scenario == 'control':
            gc_conf = get_config('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
            gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')
            server = ServerWithGCI(env, i, server_conf, gc_conf, gci_conf, log_path)

        elif scenario == 'baseline':
            gc_conf = get_config('config/gc.ini', 'gc sleep_time-0.00001 threshold-0.75') # Should change some configuration
            server = Server(env, i, server_conf, gc_conf, log_path)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    requests_conf = get_config('config/request.ini', 'request service_time-0.006 memory-0.001606664')
    clients = Clients(env, load_balancer, client_conf, requests_conf, log_path)

    for time_stamp in range(1, int(SIM_DURATION_SECONDS + 1)):
        env.run(until=time_stamp)
        log_server_metrics(time_stamp, results_path, servers[0], scenario, load)
        log_latency(time_stamp, results_path, clients.requests, scenario, load)
        log_time_in_server(time_stamp, results_path, clients.requests, scenario, load)

    log_request(clients.requests, results_path, scenario, load)

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()