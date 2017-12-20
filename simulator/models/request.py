from log import get_logger
import simpy

class Request(object):

    def __init__(self, env, id, client, load_balancer, conf, log_path=None):
        self.id = id
        self.env = env
        self.created_at = self.env.now
        self.client = client
        self.load_balancer = load_balancer

        self.service_time = float(conf['service_time'])

        self._sent_time = None
        self._arrived_time = None
        self._finished_time = None
        self._latency_time = None
        
        # Version 0.0.0 metrics
        self._interrupted_time = None
        self.redirects = 0

    def run(self):
        yield self.env.timeout(0)
    
    def interrupted_at(self, time):
        self._interrupted_time = time
        yield self.env.timeout(0)
        
    def redirected(self):
        self.redirects += 1
        yield self.env.timeout(0)

    def sent_at(self):
        self._sent_time = self.env.now
        yield self.env.timeout(0)

    def arrived_at(self):
        self._arrived_time = self.env.now
        yield self.env.timeout(0)

    def finished_at(self):
        self._finished_time = self.env.now
        self._latency_time = self._finished_time - self._sent_time
        yield self.env.timeout(0)
        