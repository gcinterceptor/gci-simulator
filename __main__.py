from models import LoadBalancer, ServerControl, ServerBaseline
from config import get_config
from log import log_request, log_gc
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

    env = simpy.Environment()

    requests_conf = get_config('config/request.ini', 'request memory-0.001606664')
    if load == 'high':
        server_conf = get_config('config/server.ini', "server high")
        loadbalancer_conf = get_config('config/loadbalancer.ini', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')

    elif load == 'low':
        server_conf = get_config('config/server.ini', "server low")
        loadbalancer_conf = get_config('config/loadbalancer.ini', 'clientlb sleep_time-0.00001 create_request_rate-35 max_requests-inf')

    else:
        raise Exception("INVALID LOAD")

    servers = list()
    load_balancer = LoadBalancer(loadbalancer_conf, requests_conf, env)
    for i in range(NUMBER_OF_SERVERS):
        if scenario == 'control':
            gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-10 initial_get-0.9')
            server = ServerControl(server_conf, gci_conf, env, i)

        elif scenario == 'baseline':
            server = ServerBaseline(server_conf, env, i)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    before_count = NUMBER_OF_SERVERS * [0]
    before_time = NUMBER_OF_SERVERS * [0]
    gc_count = list()  # Number of GC executions per second.
    gc_time = list()  # Amount of time collecting garbage per second.
    for until in range(1, int(SIM_DURATION_SECONDS) + 1):
        env.run(until=until)

        index = 0
        for server in servers:
            gc_count.append(server.collects_performed - before_count[index])
            gc_time.append(server.gc_exec_time_sum - before_time[index])
            before_count[index] = server.collects_performed
            before_time[index] = server.gc_exec_time_sum
            index += 1

    if SIM_DURATION_SECONDS > int(SIM_DURATION_SECONDS): # it means that it is a float indeed.
        env.run(until=SIM_DURATION_SECONDS) # We must run all duration

    log_request(load_balancer.requests, results_path, scenario, load)
    log_gc(gc_count, gc_time, results_path, scenario, load)

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()