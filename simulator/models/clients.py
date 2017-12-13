from log import get_logger
import simpy

class Request(object):

    def __init__(self, id, created_at, client, load_balancer, conf, log_path=None):
        self.id = id
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
        if self.logger:
            self.logger.info(" The Request %d was interrupted for %.3f" % (self.id, time))
        self._interrupted_time = time
        
    def redirected(self):
        self.redirects += 1
        if self.logger:
            self.logger.info(" The Request %d was redirect %d times" % (self.id, self.redirects))

    def sent_at(self, time):
        self._sent_time = time

    def arrived_at(self, time):
        self._arrived_time = time

    def finished_at(self, time):
        self._finished_time = time
        self._latency_time = self._finished_time - self._sent_time
        
        if self.logger:
            self.logger.info(" At %.3f, The Request %d was finished. Latency: %.3f" % (self.id, self._finished_time, self._latency_time))

class Clients(object):

    def __init__(self, env, server, conf, requests_conf, log_path=None):
        self.env = env
        self.server = server
        self.requests = list()
        self.sleep_time = 1 / int(conf['create_request_rate'])

        if log_path:
            self.logger = get_logger(log_path + "/clients.log", "CLIENTS")
        else:
            self.logger = None

        self.create_request = env.process(self.create_requests(int(conf['create_request_rate']),float(conf['max_requests']), requests_conf, log_path))

    def send_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The Request %d was send to Load Balancer" % (self.env.now, request.id))
            
        yield self.env.process(self.server.request_arrived(request))

    def create_requests(self, create_request_time, max_requests, requests_conf, log_path):
        count_requests = 0
        while count_requests <= max_requests:
            count_requests += 1
            
            if self.logger:
                self.logger.info(" At %.3f, Created Request with id %d" % (self.env.now, count_requests))
                
            request = Request(count_requests, self.env.now, self, self.server, requests_conf, log_path)
            yield self.env.process(self.send_request(request))
            yield self.env.timeout(self.sleep_time)

    def success_request(self, request):
        request.finished_at(self.env.now)
        self.requests.append(request)
        yield self.env.timeout(0)

