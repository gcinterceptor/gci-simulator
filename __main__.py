from models import ClientLB, ServerControl, ServerBaseline
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

    if len(args) >= 7:
        log_path = args[6]
        create_directory(log_path)
    else:
        log_path = None

    requests_conf = get_config('config/request.ini', 'request service_time-0.006 memory-0.001606664')
    server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')
    if load == 'high':
        delay = ' delay-0.102'
        loadbalancer_conf = get_config('config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')

    elif load == 'low':
        delay = ' delay-0.038'
        loadbalancer_conf = get_config('config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-35 max_requests-inf')

    else:
        raise Exception("INVALID LOAD")

    env = simpy.Environment()

    servers = list()
    load_balancer = ClientLB(env, loadbalancer_conf, requests_conf, log_path)
    for i in range(NUMBER_OF_SERVERS):
        if scenario == 'control':
            if load == 'high':
                collect_duration = ' collect_duration-0.151'
            else:
                collect_duration = ' collect_duration-0.308'

            gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.9' + collect_duration + delay)
            gci_conf = get_config('config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-10 initial_eget-0.9')
            server = ServerControl(env, i, server_conf, gc_conf, gci_conf, log_path)

        elif scenario == 'baseline':
            if load == 'high':
                collect_duration = ' collect_duration-0.019583333333333332'
            else:
                collect_duration = ' collect_duration-0.019333333333333332'

            gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.75' + collect_duration + delay)
            server = ServerBaseline(env, i, server_conf, gc_conf, log_path)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    before_count = NUMBER_OF_SERVERS * [0]
    before_time = NUMBER_OF_SERVERS * [0]
    gc_count = list()  # how much gc executions was runned per second.
    gc_time = list()  # how much time was used gcing per second.
    for until in range(1, int(SIM_DURATION_SECONDS) + 1):
        env.run(until=until)

        index = 0
        for server in servers:
            gc_count.append(server.gc.collects_performed - before_count[index])
            gc_time.append(server.gc.gc_exec_time_sum - before_time[index])
            before_count[index] = server.gc.collects_performed
            before_time[index] = server.gc.gc_exec_time_sum
            index += 1

    if SIM_DURATION_SECONDS > int(SIM_DURATION_SECONDS): # it means that it is a float indeed.
        env.run(until=SIM_DURATION_SECONDS) # We must run all duration

    log_request(load_balancer.requests, results_path, scenario, load)
    log_gc(gc_count, gc_time, results_path, scenario, load)

    after = time.time()
    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()