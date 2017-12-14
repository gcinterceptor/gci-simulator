from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from config import get_config
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

    if len(args) >= 6:
        results_path = args[5]
    else:
        results_path = "results"

    if len(args) >= 7:
        log_path = args[6]
    else:
        log_path = None

    create_directory(results_path)
    if log_path:
        create_directory(log_path)

    if load == 'high':
        server_conf = get_config('config/server.ini', 'server high_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-150 max_requests-inf')

    elif load == 'low':
        server_conf = get_config('config/server.ini', 'server low_load')
        client_conf = get_config('config/clients.ini', 'clients create_request_rate-35 max_requests-inf')
    
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.00001')
    
    env = simpy.Environment()
    
    load_balancer = LoadBalancer(env, loadbalancer_conf, log_path)

    servers = list()
    for i in range(SERVERS_NUMBER):
        if scenario == 'control':
            server = ServerWithGCI(env, i, server_conf, log_path)

        elif scenario == 'baseline':
            server = Server(env, i, server_conf, log_path)

        load_balancer.add_server(server)
        servers.append(server)

    requests_conf = get_config('config/request.ini', 'request service_time-0.006 memory-0.001606664')
    clients = Clients(env, load_balancer, client_conf, requests_conf, log_path)

    for time_stamp in range(1, int(SIM_DURATION_SECONDS + 1)):
        env.run(until=time_stamp)
    
    for request in clients.requests:
        print((request._latency_time, request.redirects))

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()
    
