from log import get_logger
import simpy

class Server(object):

    def __init__(self, env, id, conf, log_path, shedded_requests_rate, seed):
        self.env = env
        self.id = id
        
        if log_path:
            self.logger = get_logger(log_path + "/server.log", "SERVER")
        else:
            self.logger = None
        
        self.processed_requests = 0
        self.processed_requests_max = 1 / shedded_requests_rate
        
        self.times_interrupted = 0
        
    def request_arrived(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d arrived at Server %d" % (self.env.now, request.id, self.id))
        
        request.arrived_at()
        
        if(self.processed_requests < self.processed_requests_max):
            self.processed_requests += 1
            self.env.process(self.process_request(request))
        else:
            self.interrupt_request(request)
            self.processed_requests = 0
    
    def process_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was processed at Server %d" % (self.env.now, request.id, self.id))
        
        request.processed_at(self.id)
        yield self.env.timeout(request.cpu_time)
        
        self.env.process(request.load_balancer.success_request(request))
        
    def interrupt_request(self, request):
        if self.logger:
            self.logger.info(" At %.3f, The request %d was interrupted at Server %d" % (self.env.now, request.id, self.id))

        self.env.process(request.load_balancer.shed_request(request, self))

