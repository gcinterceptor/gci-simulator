from models import ClientLB, ServerControl, ServerBaseline
from config import get_config
from log import log_request
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

    requests_conf = get_config('config/request.ini', 'request service_time-0.006 memory-0.001606664')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')
    if load == 'high':
        loadbalancer_conf = get_config('config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')

    elif load == 'low':
        loadbalancer_conf = get_config('config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-35 max_requests-inf')

    else:
        raise Exception("INVALID LOAD")

    env = simpy.Environment()

    servers = list()
    load_balancer = ClientLB(env, loadbalancer_conf, requests_conf, log_path)
    for i in range(NUMBER_OF_SERVERS):
        if scenario == 'control':
            gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.9 collect_duration-0.0019 delay-1')
            gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-10 initial_eget-0.9')
            server = ServerControl(env, i, server_conf, gc_conf, gci_conf, log_path)

        elif scenario == 'baseline':
            gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.75 collect_duration-0.0019 delay-1')
            server = ServerBaseline(env, i, server_conf, gc_conf, log_path)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    env.run(until=SIM_DURATION_SECONDS)
    log_request(load_balancer.requests, results_path, scenario, load)

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))

if __name__ == '__main__':
    main()