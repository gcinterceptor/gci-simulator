from log import get_logger
import simpy

class Request(object):

    def __init__(self, env, id, client, load_balancer, communication_time, log_path=None):
        self.id = id
        self.env = env
        self.client = client
        self.load_balancer = load_balancer
        self.cpu_time = 0
        self.communication_time = communication_time
        self.server_id = None
        self.redirects = 0

        self._sent_time = None
        self._arrived_time = None
        self._processed_time = None
        self._finished_time = None
        self._latency_time = None
        
    def redirected(self):
        self.redirects += 1

    def sent_at(self):
        self._sent_time = self.env.now

    def arrived_at(self):
        self._arrived_time = self.env.now

    def processed_at(self, server_id):
        self._processed_time = self.env.now
        self.server_id = server_id

    def finished_at(self):
        self._finished_time = self.env.now
        self._latency_time = self._finished_time - self._sent_time
        