from models import LoadBalancer, Server
from log import log_request
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
    env = simpy.Environment()

    LOAD = int(env_var['LOAD'])
    if LOAD > 0:
        loadbalancer_load = LOAD

    else:
        raise Exception("INVALID LOAD")

    itr = 0
    servers = list()
    load_balancer = LoadBalancer(loadbalancer_load, env)
    DATA_PATH = env_var['DATA_PATH']
    INPUT_FILE_NAMES = env_var['INPUT_FILE_NAMES'].split()
    NUMBER_OF_SERVERS = int(env_var['NUMBER_OF_SERVERS'])
    for i in range(NUMBER_OF_SERVERS):
        data = build_data(DATA_PATH + INPUT_FILE_NAMES[itr])
        itr = (itr + 1) % len(INPUT_FILE_NAMES)
        server = Server(env, i, data)
        load_balancer.add_server(server)
        servers.append(server)

    DURATION = float(env_var['DURATION'])
    env.run(until=DURATION)
    after = time.time()

    RESULTS_PATH = env_var['RESULTS_PATH']
    RESULTS_NAME = env_var['RESULTS_NAME']
    log_request(load_balancer.requests, RESULTS_PATH, RESULTS_NAME, "w")

    text = "created requests: " + str(load_balancer.created_requests) \
           + "\nshedded requests: " + str(load_balancer.shedded_requests) \
           + "\nlost requests: " + str(load_balancer.lost_requests) \
           + "\nsucceeded requests: " + str(load_balancer.succeeded_requests)
    print(text)

    print("Time of simulation execution in seconds: %.4f" % (after - before))


if __name__ == '__main__':
    main()