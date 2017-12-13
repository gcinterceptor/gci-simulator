from simulator.models import Clients, LoadBalancer, ServerWithGCI, Server
from log import iniciate_csv_files, csv_writer
from config import get_config
import simpy, os, time, math

def main():
    before = time.time()

    log_path = 'logs'
    results_path = "results"
    env = simpy.Environment()

    client_conf = get_config('config/clients.ini', 'clients sleep_time-0.00001 create_request_rate-35 max_requests-inf')
    loadbalancer_conf = get_config('config/loadbalancer.ini', 'loadbalancer sleep_time-0.0035')
    requests_conf = get_config('config/request.ini', 'request service_time-0.0035 memory-0.02')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

    server_control = ServerWithGCI(env, 1, server_conf, log_path)
    load_balancer_control = LoadBalancer(env, server_control, loadbalancer_conf, log_path)
    clients_control = Clients(env, load_balancer_control, client_conf, requests_conf, log_path)

    server_baseline = Server(env, 1, server_conf, log_path)
    load_balancer_baseline = LoadBalancer(env, server_baseline, loadbalancer_conf, log_path)
    clients_baseline = Clients(env, load_balancer_baseline, client_conf, requests_conf, log_path)

    SIM_DURATION_SECONDS = 1
    env.run(until=SIM_DURATION_SECONDS)

    after = time.time()
    print("Time of execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()