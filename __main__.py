from simulator.models import Clients, LoadBalancer, Server
from config import get_config
from log import _initiate_csv_files
from metrics import log_request
import simpy, os, sys, time

def create_directory(path):
    if not os.path.isdir(path):
        os.mkdir(path)

def main():
    before = time.time()

    args = sys.argv
    SERVERS_NUMBER = int(args[1])
    SIM_DURATION_SECONDS = float(args[2])
    load = args[3]
    shedded_requests_rate = float(args[4])
    network_communication_time = float(args[5])
    requests_cpu_time = float(args[6])

    if len(args) >= 8:
        results_path = args[7]
    else:
        results_path = "results"

    if len(args) >= 9:
        seed = int(args[8])
    else:
        seed = int(time.time())

    if len(args) >= 10:
        log_path = args[9]
    else:
        log_path = None

    create_directory(results_path)
    _initiate_csv_files(results_path, SERVERS_NUMBER, load, shedded_requests_rate, network_communication_time, requests_cpu_time)
    
    if log_path:
        create_directory(log_path)

    if load == 'high':
        server_conf = get_config('config/server.ini', 'server high_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-150 max_requests-inf')

    elif load == 'low':
        server_conf = get_config('config/server.ini', 'server low_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-35 max_requests-inf')
    
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'DEFAULT')
    
    env = simpy.Environment()

    load_balancer = LoadBalancer(env, loadbalancer_conf, network_communication_time, log_path)

    servers = list()
    for i in range(SERVERS_NUMBER):
        server = Server(env, i, server_conf, log_path, shedded_requests_rate, seed)

        load_balancer.add_server(server)
        servers.append(server)

    clients = Clients(env, load_balancer, client_conf, SERVERS_NUMBER, SIM_DURATION_SECONDS, requests_cpu_time, log_path)

    env.run(until=int(SIM_DURATION_SECONDS + 1))
        
    log_request(clients.requests, results_path, SERVERS_NUMBER, load, shedded_requests_rate, network_communication_time, requests_cpu_time)
    
    after = time.time()
    
    print("Used seed to random numbers generator: %d" % seed)
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()
    
