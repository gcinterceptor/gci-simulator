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

    env_var = os.environ
    NUMBER_OF_SERVERS = int(env_var['NUMBER_OF_SERVERS'] )
    DURATION = float(env_var['DURATION'] )
    SCENARIO = env_var['SCENARIO']
    LOAD = int(env_var['LOAD'] )
    RESULTS_PATH = env_var['RESULTS_PATH']
    create_directory(RESULTS_PATH)

    DATA_PATH = env_var['DATA_PATH']
    SERVICE_TIME_FILE_NAME = env_var['SERVICE_TIME_FILE_NAME']
    SERVICE_TIME_DATA_COLUMN = int(env_var['SERVICE_TIME_DATA_COLUMN'] )
    service_time_data = build_data(DATA_PATH + SERVICE_TIME_FILE_NAME, SERVICE_TIME_DATA_COLUMN)

    env = simpy.Environment()

    if LOAD > 0:
        loadbalancer_load = LOAD

    else:
        raise Exception("INVALID LOAD")

    servers = list()
    load_balancer = LoadBalancer(loadbalancer_load, env)
    for i in range(NUMBER_OF_SERVERS):
        if SCENARIO == 'control':
            PROCESSED_REQUESTS_FILE_NAME = env_var['SHEDDING_FILE_NAME']
            PROCESSED_REQUESTS_COLUMN = 0
            NUMBER_OF_FILES = int(env_var['SHEDDING_NUMBER_OF_FILES'])
            processed_requests_data = build_data(DATA_PATH + PROCESSED_REQUESTS_FILE_NAME + str((i % NUMBER_OF_FILES) + 1), PROCESSED_REQUESTS_COLUMN)

            SHEDDED_REQUESTS_FILE_NAME = env_var['SHEDDING_FILE_NAME']
            SHEDDED_REQUESTS_COLUMN = 1
            shedded_requests_data = build_data(DATA_PATH + SHEDDED_REQUESTS_FILE_NAME + str((i % NUMBER_OF_FILES) + 1), SHEDDED_REQUESTS_COLUMN)

            server = ServerControl(env, i, service_time_data, processed_requests_data, shedded_requests_data)

        elif SCENARIO == 'baseline':
            server = ServerBaseline(env, i, service_time_data)

        else:
            raise Exception("INVALID SCENARIO")

        load_balancer.add_server(server)
        servers.append(server)

    env.run(until=DURATION) # We must run all duration

    after = time.time()

    RESULTS_NAME = env_var['RESULTS_NAME']
    log_request(load_balancer.requests, RESULTS_PATH, RESULTS_NAME)

    text = "created requests: " + str(load_balancer.created_requests) \
           + "\nshedded requests: " + str(load_balancer.shedded_requests) \
           + "\nlost requests: " + str(load_balancer.lost_requests) \
           + "\nsucceeded requests: " + str(load_balancer.succeeded_requests)
    print(text)

    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()