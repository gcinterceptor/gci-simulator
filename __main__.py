from models import LoadBalancer, ServerControl, ServerBaseline
from log import log_request, txt_writer
import simpy, os, sys, time


def create_directory(path):
    if not os.path.isdir(path):
        os.mkdir(path)


def build_data(file_name, column):
    data = list()
    file_obj = open(file_name + ".csv", 'r')
    for linha in file_obj:
        try:
            result = list(map(float, linha.split(",")))
            data.append(result[column])
        except:
            # avoid invalid values.
            pass

    return data


def main():
    before = time.time()

    args = sys.argv

    NUMBER_OF_SERVERS = int(args[1])
    SIM_DURATION_SECONDS = float(args[2])
    SCENARIO = args[3]
    LOAD = int(args[4])
    RESULTS_PATH = args[5]
    create_directory(RESULTS_PATH)

    DATA_PATH = args[6]
    SERVICE_TIME_FILE_NAME = args[7]
    SERVICE_TIME_COLUMN = int(args[8])
    service_time_data = build_data(DATA_PATH + SERVICE_TIME_FILE_NAME, SERVICE_TIME_COLUMN)

    env = simpy.Environment()

    if LOAD > 0:
        loadbalancer_load = LOAD

    else:
        raise Exception("INVALID LOAD")

    servers = list()
    load_balancer = LoadBalancer(loadbalancer_load, env)
    for i in range(NUMBER_OF_SERVERS):
        if SCENARIO == 'control':
            PROCESSED_REQUESTS_FILE_NAME = args[10]
            PROCESSED_REQUESTS_COLUMN = 0
            processed_requests_data = build_data(DATA_PATH + PROCESSED_REQUESTS_FILE_NAME + str((i % 4) + 1), PROCESSED_REQUESTS_COLUMN)

            SHEDDED_REQUESTS_FILE_NAME = args[10]
            SHEDDED_REQUESTS_COLUMN = 1
            shedded_requests_data = build_data(DATA_PATH + SHEDDED_REQUESTS_FILE_NAME + str((i % 4) + 1), SHEDDED_REQUESTS_COLUMN)

            server = ServerControl(env, i, service_time_data, processed_requests_data, shedded_requests_data)

        elif SCENARIO == 'baseline':
            server = ServerBaseline(env, i, service_time_data)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    for until in range(1, int(SIM_DURATION_SECONDS) + 1):
        env.run(until=until)

    if SIM_DURATION_SECONDS > int(SIM_DURATION_SECONDS): # it means that it is a float indeed.
        env.run(until=SIM_DURATION_SECONDS) # We must run all duration

    for request in load_balancer.requests:
        request._latency = request._latency

    EXPERIMENT_NUMBER = args[9]
    log_request(load_balancer.requests, RESULTS_PATH, str(NUMBER_OF_SERVERS) + "instances_loadbalancer_request_status", SCENARIO, str(LOAD) + "_" + EXPERIMENT_NUMBER)

    after = time.time()

    text = "number of service instances: " + str(NUMBER_OF_SERVERS) \
           + "\nsimulation time duration: " + str(SIM_DURATION_SECONDS) \
           + "\nconfiguration scenario: " + SCENARIO \
           + "\nworkload per instance: " + str(LOAD) + "req/sec" \
           + "\ngeneral workload: " + str(LOAD * NUMBER_OF_SERVERS) + "req/sec" \
           + "\ncreated requests: " + str(load_balancer.created_requests) \
           + "\nshedded requests: " + str(load_balancer.shedded_requests) \
           + "\nlost requests: " + str(load_balancer.lost_requests) \
           + "\nsucceeded requests: " + str(load_balancer.succeeded_requests)
    txt_writer(RESULTS_PATH + "/simulation_info_" + SCENARIO + "_load_" + str(LOAD) + ".txt", text)

    print("created requests: %f" % load_balancer.created_requests)
    print("shedded requests: %f" % load_balancer.shedded_requests)
    print("lost requests: %f" % load_balancer.lost_requests)
    print("succeeded requests: %f" % load_balancer.succeeded_requests)
    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()