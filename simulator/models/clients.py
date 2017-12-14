from log import get_logger
from .request import Request
import simpy

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
                
            request = Request(self.env, count_requests, self, self.server, requests_conf, log_path)
            yield self.env.process(self.send_request(request))
            yield self.env.timeout(self.sleep_time)

    def success_request(self, request):
        if self.logger:
                self.logger.info(" At %.3f, Request with id %d was processed with Success" % (self.env.now, request.id))
        
        self.requests.append(request)
        yield self.env.process(request.finished_at())
        yield self.env.timeout(0)

