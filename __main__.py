from models import LoadBalancer, Server
from log import log_request, log_debbug
import simpy, os, time


def create_directory(path):
    if not os.path.isdir(path):
        os.mkdir(path)


def build_data(file_name):
    data = list()
    file_obj = open(file_name + ".csv", 'r')
    for linha in file_obj:
        try:
            result = list(map(float, linha.split(",")))
            result.pop(0)  # removes the timestamp column
            data.append(result)
        except:
            # avoid invalid values.
            pass

    return data


def main():
    before = time.time()

    env_var = os.environ
    NUMBER_OF_SERVERS = int(env_var['NUMBER_OF_SERVERS'])
    DURATION = float(env_var['DURATION'])
    LOAD = int(env_var['LOAD'] )
    RESULTS_PATH = env_var['RESULTS_PATH']
    create_directory(RESULTS_PATH)

    DATA_PATH = env_var['DATA_PATH']
    SERVICE_TIME_FILE_NAME = env_var['SERVICE_TIME_FILE_NAME']
    data = build_data(DATA_PATH + SERVICE_TIME_FILE_NAME)

    env = simpy.Environment()

    if LOAD > 0:
        loadbalancer_load = LOAD

    else:
        raise Exception("INVALID LOAD")

    servers = list()
    load_balancer = LoadBalancer(loadbalancer_load, env)
    for i in range(NUMBER_OF_SERVERS):
        server = Server(env, i, data)
        load_balancer.add_server(server)
        servers.append(server)

    env.run(until=DURATION)
    after = time.time()

    RESULTS_NAME = env_var['RESULTS_NAME']
    log_request(load_balancer.requests, RESULTS_PATH, RESULTS_NAME, "w")

    if env_var['DEBBUG'].upper() == "TRUE":
        log_debbug(load_balancer.requests, RESULTS_PATH + "debbug/", RESULTS_NAME, "w")

    text = "created requests: " + str(load_balancer.created_requests) \
           + "\nshedded requests: " + str(load_balancer.shedded_requests) \
           + "\nlost requests: " + str(load_balancer.lost_requests) \
           + "\nsucceeded requests: " + str(load_balancer.succeeded_requests)
    print(text)

    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()