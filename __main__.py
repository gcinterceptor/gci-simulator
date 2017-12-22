from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
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
    scenario = args[3]
    load = args[4]
    availability_rate = float(args[5])

    if len(args) >= 7:
        results_path = args[6]
    else:
        results_path = "results"

    if len(args) >= 8:
        log_path = args[7]
    else:
        log_path = None

    create_directory(results_path)
    _initiate_csv_files(results_path, SERVERS_NUMBER, scenario, load, availability_rate)
    
    if log_path:
        create_directory(log_path)

    if load == 'high':
        server_conf = get_config('config/server.ini', 'server high_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-150 max_requests-inf')

    elif load == 'low':
        server_conf = get_config('config/server.ini', 'server low_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-35 max_requests-inf')
    
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.001')
    
    env = simpy.Environment()
    
    load_balancer = LoadBalancer(env, loadbalancer_conf, log_path)

    avg_unavailable_time = 1.00
    avg_available_time = availability_rate * avg_unavailable_time
    
    servers = list()
    for i in range(SERVERS_NUMBER):
        if scenario == 'control':
            server = ServerWithGCI(env, i, server_conf, log_path, avg_available_time, avg_unavailable_time)

        elif scenario == 'baseline':
            server = Server(env, i, server_conf, log_path, avg_available_time, avg_available_time)

        load_balancer.add_server(server)
        servers.append(server)

    clients = Clients(env, load_balancer, client_conf, SERVERS_NUMBER, log_path)

    for time_stamp in range(1, int(SIM_DURATION_SECONDS + 1)):
        env.run(until=time_stamp)
        
    log_request(clients.requests, results_path, SERVERS_NUMBER, scenario, load, availability_rate)
    
    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()
    
