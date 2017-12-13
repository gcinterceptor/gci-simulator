from log import get_logger
import simpy

class Request(object):

    def __init__(self, created_at, client, load_balancer, conf, log_path=None):
        self.created_at = created_at
        self.client = client
        self.load_balancer = load_balancer

        self.service_time = float(conf['service_time'])

        self.done = False
        self._sent_time = None
        self._arrived_time = None
        self._finished_time = None
        self._latency_time = None
        
        # Version 0.0.0 metrics
        self._interrupted_time = None
        self.redirects = 0

        if log_path:
            self.logger = get_logger(log_path + "/request.log", "REQUEST")
        else:
            self.logger = None

    def run(self, env):
        yield env.timeout(0)
    
    def interrupted_at(self, time):
        self._interrupted_time = time
        
    def redirected(self):
        self.redirects += 1

    def sent_at(self, time):
        self._sent_time = time

    def arrived_at(self, time):
        self._arrived_time = time

    def finished_at(self, time):
        self._finished_time = time
        self._latency_time = self._finished_time - self._sent_time
        
        if self.logger:
            self.logger.info(" At %.3f, Request was finished. Latency: %.3f" % (self._finished_time, self._latency_time))

class Clients(object):

    def __init__(self, env, server, conf, requests_conf, log_path=None):
        self.env = env
        self.server = server
        self.requests = list()
        self.sleep_time = float(conf['sleep_time'])

        self.queue = simpy.Store(env)               # the queue of requests

        if log_path:
            self.logger = get_logger(log_path + "/clients.log", "CLIENTS")
        else:
            self.logger = None

        self.action = env.process(self.send_requests())
        self.create_request = env.process(self.create_requests(int(conf['create_request_rate']),float(conf['max_requests']), requests_conf, log_path))

    def send_requests(self):
        while True:
            if len(self.queue.items) > 0:           # check if there is any request to be processed
                request = yield self.queue.get()    # get a request from store
                yield self.env.process(self.server.request_arrived(request))

            yield self.env.timeout(self.sleep_time) # wait for...

    def create_requests(self, create_request_time, max_requests, requests_conf, log_path):
        count_requests = 1
        while count_requests <= max_requests:
            request = Request(self.env.now, self, self.server, requests_conf, log_path)
            yield self.queue.put(request)
            yield self.env.timeout(1.0 / create_request_time)
            count_requests += 1

    def success_request(self, request):
        request.finished_at(self.env.now)
        self.requests.append(request)
        yield self.env.timeout(0)

