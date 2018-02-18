from log import get_logger
from .request import Request
import simpy

class Clients(object):

    def __init__(self, env, server, conf, servers_number, create_requests_until, requests_cpu_time, log_path=None):
        self.env = env
        self.server = server
        self.requests = list()
        
        self.sleep_time = 1 / (int(conf['create_request_rate']) * servers_number)
        self.create_requests_until = create_requests_until
        self.requests_cpu_time = requests_cpu_time

        if log_path:
            self.logger = get_logger(log_path + "/clients.log", "CLIENTS")
        else:
            self.logger = None

        self.create_request = env.process(self.create_requests(int(conf['create_request_rate']), log_path))

    def send_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The Request %d was send to Load Balancer" % (self.env.now, request.id))
            
        self.server.request_arrived(request)

    def create_requests(self, create_request_time, log_path):
        request_id = 1
        while self.env.now < self.create_requests_until:
            if self.logger:
                self.logger.info(" At %.3f, Created Request with id %d" % (self.env.now, request_id))
                
            request = Request(self.env, request_id, self, self.server, self.requests_cpu_time, log_path)
            
            self.send_request(request)
            yield self.env.timeout(self.sleep_time)
            
            request_id += 1

    def success_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, Request with id %d was processed with Success" % (self.env.now, request.id))
        
        request.finished_at()
        self.requests.append(request)
        
    def refuse_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, Request with id %d was refused" % (self.env.now, request.id))
        
        request.request_refused()
        self.requests.append(request)
        
